# r53

```
A tool that facilitates interactions with Route 53

Usage:
  r53 [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  update      Update a Route 53 record with the system's public IP

Flags:
  -h, --help          help for r53
  -z, --zone string   hosted zone ID

Use "r53 [command] --help" for more information about a command.
```

## update

```
Update a Route 53 record with the system's public IP

Usage:
  r53 update [flags]

Flags:
  -h, --help            help for update
  -r, --record string   record name
  -t, --ttl int         record TTL (default 300)

Global Flags:
  -z, --zone string   hosted zone ID
```
