# Marathon Client

A client to interact with a marathon server.  It will deploy applications, delete applications, track the progress of a deployment, and report back at the end of the job.

It will automatically detect if you're deploying an application or a group, and if the deployment is an update or a new job.

## Usage

Usage of ./marathon-client:

| Flag | Description  |
|------|--------------|
| -d   | Debug output |
| -f   | Job file     | 
| -m   | Marathon URL |
| -u   | Username for basic auth |
| -p   | Password for basic auth |
| -force | Force deploy over existing deployment |
| -delete | Delete an existing application |

Note that Job file can be set to "-" to read from STDIN.

Examples:
```
# Deploy
marathon-client -f job.json -m marathon.mydomain:8080 -u user -p pass
cat job.json | marathon-client -f - -m marathon.mydomain:8080

# Delete
echo '{"id": "/service-name"}' | marathon-client -m http://marathon.url --delete -u user -p pass -f -
```

## Compatibility

This requires marathon 0.9.0 or later.
Most testing has been done on the 0.11.x tree, but anything after 0.9.0 is supported and should work.
