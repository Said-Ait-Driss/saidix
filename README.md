# saidis

<p>
saidis is an in-momery database like redis. built on go
</p>
<strong>
<a href="https://www.databass.dev/"> referer followed toturial</a>
</strong>
# How
-  <strong>RESP</strong> protocol as a tool which allows the server to receive commands and respond with responses.
-  Use <strong>go routines</strong> to handle multiple connections simultaneously.
- Write data to disk using the Append Only File <strong>(AOF)</strong>, which is one of the methods Redis uses for persistence. This way, if the server crashes or restarts, we can restore the data.
<img src="https://www.build-redis-from-scratch.dev/images/diagram.svg">
## data presistance :
<p>
    generally they are two methods to presistance data in the database world :
        either RDB or AOF

        + RDB (Redis Database) : a snapshot of the data that is created at regular intervals.
        + AOF (Append Only File) : records each command in the file as RESP. reads all the RESP commands from the AOF file and executes them in memory ( in case a restart occurs ).

         -- format of AOF file after executing the two commands :
</p>

 ```ruby
            set name said
            set email "saidaitdrissofficial@gmail.com"
```
        <p>The content of the file will be :</p>

```ruby
*2
$3
set
$4
name
*3
$3
set
$4
name
$4
said
*3
$3
set
$5
email
$30
saidaitdrissofficial@gmail.com
```


## note : still using redis client (redis-cli) to test this server 
## next steps :
### support more commands : https://redis.io/docs/latest/commands/ 
### is to build saidis client to interact with saidis server
## add some examples of : Pipelines for batching commands
### add Pub/Sub for sending messages between users