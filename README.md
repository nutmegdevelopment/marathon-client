# Marathon Client

A client to interact with a marathon server.  It will deploy applications, track the progress of a deployment, and report back at the end of the job.

Usage of ./marathon-client:

| Flag | Description  |
|------|--------------|
| -d   | Debug output |
| -f   | Job file     | 
| -m   | Marathon URL |

Note that Job file can be set to "-" to read from STDIN.