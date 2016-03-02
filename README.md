# mrelay
A deamon that forwards a message from localhost (listening on UDP or TCP) to remote host on an arbitraty protocol (*pluggable*). 

Currently supported forwarders:
  * HTTP
  * Kafka
  * TCP
  * UDP
  
Message format is user-defined.
