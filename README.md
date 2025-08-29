Implementation of network device discovery using raw ICMP sockets (RFC 792 Echo).

```
go build -o main main.go
sudo ./main
# View devices found during network scan:
sleep 10 && cat /proc/net/arp
```