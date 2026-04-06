# Ownapi
A local-first system for setting up pipelines to run custom logic and get quick feedback in the browser. I'm currently developing it for 2 use cases:
1. A dashboard for my Kindle that will show public transport and weather information updated every 1 minute
2. Automated ingestion of activities from my Garmin Forerunner to generate graphs per workout (e.g. HR, average pace, cadence, etc) and PB's across workouts (e.g. fastest 1km, fastest cadence, etc)

It's currently in an experimental state with limited functionality including:
- 2 node types: HTTP (for making HTTP requests) and JSON (for extracting information from JSON)
- a data type for defining pipelines comprising of multiple stages, where all nodes within a stage can run concurrently and data is passed between stages using a map
- a system for identifying secrets in the `_config/pipelines.json` file in the format `secret:NAME`, creating a local encrypted file to store them, prompting the user to enter them if they're not defined and loading them
    into the Linux Kernel Key Retention Service at runtime. NOTE: Currently only supported for some fields in the HTTP node
- a frontend web gui that opens a websocket per client to show live updates for pipeline executions as they complete server-side (currently sends updates every 10us, may change to Server Sent Events in the future)

# Setup
## Fedora
To setup libusb for development install `libusb1-devel`

1. `sudo nano /etc/udev/rules.d/51-garmin.rules`
2. Paste: `SUBSYSTEM=="usb", ATTR{idVendor}=="091e", MODE="0666"`
3. `sudo udevadm control --reload-rules`

Follow this guide to disable auto mount: https://discussion.fedoraproject.org/t/help-disabling-automount-of-specific-media/71295/4