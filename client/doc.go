// The client is the primary interface for redis. You must first create a client with redis address for working.
//
//     c := NewClient("127.0.0.1:6380")
// or
//     c := NewClient("127.0.0.1:6380",DialWriteTimeout(1024),DialPassword("1234"))
//
// The most important function for client is Do function to send commands to remote server.
//
//     reply, err := c.Do("ping")
//
//     reply, err := c.Do("set", "key", "value")
//
//     reply, err := c.Do("get", "key")
//
// Connection
//
// You can use an independent connection to send commands.
//
//     //get a connection
//     conn, _ := c.Get()
//
//     //connection send command
//     conn.Do("ping")
//
// Reply Helper
//
// You can use reply helper to convert a reply to a specific type.
//
//     exists, err := Bool(c.Do("exists", "key"))
//
//     score, err := Int64(c.Do("zscore", "key", "member"))
package client
