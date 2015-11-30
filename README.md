# Marathon Client

A client to interact with a marathon server.  It will deploy applications, track the progress of a deployment, and report back at the end of the job.

Usage of ./marathon-client:

| Flag | Description  |
|------|--------------|
| -d   | Debug output |
| -f   | Job file     | 
| -m   | Marathon URL |
| -u   | Username for basic auth |
| -p   | Password for basic auth |

Note that Job file can be set to "-" to read from STDIN.

Examples:
```
marathon-client -f job.json -m marathon.mydomain:8080 -u user -p pass
cat job.json | marathon-client -f - -m marathon.mydomain:8080