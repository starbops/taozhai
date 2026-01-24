// cmd/taozhai/main.go
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/starbops/taozhai/pkg/bmc"
	"github.com/starbops/taozhai/pkg/client"
	"github.com/starbops/taozhai/pkg/discovery"
	"github.com/starbops/taozhai/pkg/inventory"
	"github.com/starbops/taozhai/pkg/version"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	// CLI flags
	kubeconfig   *string
	namespace    *string
	bmcNamespace *string
	timeout      *time.Duration
	bmcTimeout   *time.Duration
	force        *bool
	powerOff     *bool
	versionFlag  *bool
)

func init() {
	// Initialize flags
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "path to kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "path to kubeconfig file")
	}

	namespace = flag.String("namespace", "default", "namespace for discovery pods")
	bmcNamespace = flag.String("bmc-namespace", "tink-system", "namespace for BMC Job CRs")
	timeout = flag.Duration("timeout", 2*time.Minute, "timeout for pod completion")
	bmcTimeout = flag.Duration("bmc-timeout", 5*time.Minute, "timeout for BMC Job completion")
	force = flag.Bool("force", false, "skip confirmation prompt")
	powerOff = flag.Bool("power-off", false, "actually create BMC power-off Job (default is dry-run)")
	versionFlag = flag.Bool("version", false, "print version information and exit")
}

func main() {
	flag.Parse()

	// Handle --version flag
	if *versionFlag {
		fmt.Println(version.GetVersionString())
		os.Exit(0)
	}

	// Validate arguments
	args := flag.Args()
	if len(args) != 1 {
		printUsage()
		os.Exit(1)
	}
	targetIP := args[0]

	// Initialize Kubernetes client
	c, err := initializeClient(*kubeconfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing Kubernetes client: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	// Run the IP reclamation workflow
	if err := runIPReclamation(ctx, c, targetIP); err != nil {
		fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✓ Done!")
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: taozhai [options] <target-ip>\n\n")
	fmt.Fprintf(os.Stderr, "Discovers IP address hijackers and optionally powers them off via BMC.\n\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  # Dry-run mode (shows what would happen)\n")
	fmt.Fprintf(os.Stderr, "  taozhai 192.168.1.100\n\n")
	fmt.Fprintf(os.Stderr, "  # Actually power off the hijacker\n")
	fmt.Fprintf(os.Stderr, "  taozhai --power-off 192.168.1.100\n\n")
	fmt.Fprintf(os.Stderr, "  # Skip confirmation prompt\n")
	fmt.Fprintf(os.Stderr, "  taozhai --power-off --force 192.168.1.100\n")
}

func initializeClient(kubeconfigPath string) (*client.Client, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
	}

	c, err := client.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return c, nil
}

func runIPReclamation(ctx context.Context, c *client.Client, targetIP string) error {
	fmt.Println("=" + strings.Repeat("=", 60))
	fmt.Println(" Taozhai - IP Address Hijacker Detection & Reclamation")
	fmt.Println("=" + strings.Repeat("=", 60))
	fmt.Printf("\nTarget IP: %s\n", targetIP)

	// Phase 1: Discover MAC address via ARP
	fmt.Println("\n" + strings.Repeat("-", 60))
	fmt.Println("Phase 1: MAC Address Discovery")
	fmt.Println(strings.Repeat("-", 60))

	macAddr, err := discoverMAC(ctx, c, targetIP)
	if err != nil {
		return fmt.Errorf("MAC discovery failed: %w", err)
	}

	fmt.Printf("\n✓ Found MAC address: %s\n", macAddr)

	// Phase 2: Find Inventory CR by MAC address
	fmt.Println("\n" + strings.Repeat("-", 60))
	fmt.Println("Phase 2: Inventory Search")
	fmt.Println(strings.Repeat("-", 60))

	inventory, err := findInventory(ctx, c, macAddr)
	if err != nil {
		return fmt.Errorf("inventory search failed: %w", err)
	}

	inventoryName := inventory.GetName()
	fmt.Printf("\n✓ Found Inventory CR: %s\n", inventoryName)

	// Phase 3: Create BMC power-off Job (or dry-run)
	fmt.Println("\n" + strings.Repeat("-", 60))
	fmt.Println("Phase 3: BMC Power Control")
	fmt.Println(strings.Repeat("-", 60))

	if *powerOff {
		// Confirm with user unless --force is set
		if !*force {
			if !confirmPowerOff(inventoryName, macAddr, targetIP) {
				fmt.Println("\nOperation cancelled by user.")
				return nil
			}
		}

		// Actually create the BMC Job
		// Note: Machine CR has the same name as Inventory CR
		if err := createBMCJob(ctx, c, inventoryName); err != nil {
			return fmt.Errorf("BMC Job creation failed: %w", err)
		}
	} else {
		// Dry-run mode
		fmt.Println("\n[DRY-RUN MODE]")
		fmt.Printf("Would power off server: %s\n", inventoryName)
		fmt.Printf("  MAC Address: %s\n", macAddr)
		fmt.Printf("  IP Address:  %s\n", targetIP)
		fmt.Println("\nTo actually power off the server, use the --power-off flag:")
		fmt.Printf("  taozhai --power-off %s\n", targetIP)
	}

	return nil
}

func discoverMAC(ctx context.Context, c *client.Client, targetIP string) (string, error) {
	discoverer := discovery.NewDiscoverer(c).
		SetTimeout(*timeout).
		SetNamespace(*namespace)

	macAddr, err := discoverer.DiscoverMACFromIP(ctx, targetIP)
	if err != nil {
		return "", err
	}

	return macAddr, nil
}

func findInventory(ctx context.Context, c *client.Client, macAddr string) (*unstructured.Unstructured, error) {
	finder := inventory.NewFinder(c)

	// Get the Inventory CR object
	inv, err := finder.GetHardwareByMAC(ctx, macAddr)
	if err != nil {
		return nil, err
	}

	// Display formatted inventory information
	inventoryInfo, _ := finder.FindByMAC(ctx, macAddr)
	fmt.Println("\nInventory Details:")
	fmt.Println(inventoryInfo)

	return inv, nil
}

func createBMCJob(ctx context.Context, c *client.Client, inventoryName string) error {
	bmcManager := bmc.NewManager(c)

	// Machine CR name = Inventory CR name
	jobName, err := bmcManager.CreatePowerOffJob(ctx, inventoryName, *bmcNamespace)
	if err != nil {
		return err
	}

	// Wait for Job to complete
	if err := bmcManager.WaitForJobCompletion(ctx, jobName, *bmcNamespace, *bmcTimeout); err != nil {
		return fmt.Errorf("BMC Job did not complete successfully: %w", err)
	}

	fmt.Println("\n✓ Server powered off successfully")
	return nil
}

func confirmPowerOff(inventoryName, macAddr, targetIP string) bool {
	fmt.Println("\n" + strings.Repeat("!", 60))
	fmt.Println(" WARNING: This will power off the server!")
	fmt.Println(strings.Repeat("!", 60))
	fmt.Printf("\nInventory:   %s\n", inventoryName)
	fmt.Printf("MAC Address: %s\n", macAddr)
	fmt.Printf("IP Address:  %s\n", targetIP)
	fmt.Printf("\nThis action will:")
	fmt.Printf("\n  1. Create a BMC power-off Job in namespace '%s'\n", *bmcNamespace)
	fmt.Printf("  2. Power off the server to reclaim the IP address\n")
	fmt.Printf("  3. Cause service disruption if this is the wrong server\n")

	fmt.Print("\nAre you sure you want to proceed? (yes/no): ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "yes" || response == "y"
}
