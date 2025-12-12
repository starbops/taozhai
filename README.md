# taozhai

**Help! My servers just don't PXE boot!**

> I don't know who you are.  
> I don't know what you want.  
> If you're look for a ransom,  
> I can tell you I don't have IPs.  
> But what I do have, are a very particular set of skills.  
> Skills I've acquired over a very long LLM chat session.  
> Skills that make me a nightmare for servers like you.  
> If you give my IP address back now, that'll be the end of it.  
> I will not look for you.  
> I will not pursue you.  
> But if you don't, I will look for you.  
> I will find you.  
> And I will power-off you.  
> 
> – *taozhai*

## Build

```shell
make build
```

## Usage

```shell
$ ./bin/taozhai --help
Usage of ./bin/taozhai:
  -bmc-namespace string
    	namespace for BMC Job CRs (default "tink-system")
  -bmc-timeout duration
    	timeout for BMC Job completion (default 5m0s)
  -force
    	skip confirmation prompt
  -inventory-namespace string
    	namespace for Seeder Inventory CRs (default "tink-system")
  -kubeconfig string
    	path to kubeconfig file (default "/Users/starbops/.kube/config")
  -namespace string
    	namespace for discovery pods (default "default")
  -power-off
    	actually create BMC power-off Job (default is dry-run)
  -timeout duration
    	timeout for pod completion (default 2m0s)
```

Example:

```shell
$ ./bin/taozhai --kubeconfig seeder.kubeconfig --power-off 10.115.252.29
=============================================================
 Taozhai - IP Address Hijacker Detection & Reclamation
=============================================================

Target IP: 10.115.252.29

------------------------------------------------------------
Phase 1: MAC Address Discovery
------------------------------------------------------------
Creating discovery pod taozhai-discovery-1765522316...
Waiting for pod to complete...
Cleaning up pod taozhai-discovery-1765522316...

✓ Found MAC address: 52:54:00:0b:2e:01

------------------------------------------------------------
Phase 2: Inventory Search
------------------------------------------------------------

Searching for Inventory with MAC 52:54:00:0b:2e:01...

Inventory Details:
Name:      alfa-1
Namespace: tink-system
MAC:       52:54:00:0b:2e:01
Disk:      /dev/vda
Arch:      amd64


✓ Found Inventory CR: alfa-1

------------------------------------------------------------
Phase 3: BMC Power Control
------------------------------------------------------------

!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
 WARNING: This will power off the server!
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!

Inventory:   alfa-1
MAC Address: 52:54:00:0b:2e:01
IP Address:  10.115.252.29

This action will:
  1. Create a BMC power-off Job in namespace 'tink-system'
  2. Power off the server to reclaim the IP address
  3. Cause service disruption if this is the wrong server

Are you sure you want to proceed? (yes/no): yes

Creating BMC power-off Job for machine: alfa-1
✓ BMC Job created: taozhai-poweroff-alfa-1-1765522331

Waiting for BMC Job to complete (timeout: 5m0s)...
✓ BMC Job completed successfully

✓ Server powered off successfully

✓ Done!
```
