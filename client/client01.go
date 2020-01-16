package main

import (
    "crypto/tls"
    "crypto/x509"
    "fmt"
    //"io"
    "log"
    "sechat-server/lib"
)

func main() {


    frame := lib.CreateFrame([]byte("HELLO YOLO"))

    cert, err := tls.LoadX509KeyPair("certs/client.pem", "certs/client.key")
    if err != nil {
        log.Fatalf("server: loadkeys: %s", err)
    }
    config := tls.Config{
        Certificates: []tls.Certificate{cert},
        InsecureSkipVerify: true,
        CipherSuites: []uint16{tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA},
    }
    conn, err := tls.Dial("tcp", "127.0.0.1:8000", &config)
    if err != nil {
        log.Fatalf("client: dial: %s", err)
    }
    defer conn.Close()
    log.Println("client: connected to: ", conn.RemoteAddr())

    state := conn.ConnectionState()
    for _, v := range state.PeerCertificates {
        fmt.Println(x509.MarshalPKIXPublicKey(v.PublicKey))
        fmt.Println(v.Subject)
    }
    log.Println("client: handshake: ", state.HandshakeComplete)
    log.Println("client: mutual: ", state.NegotiatedProtocolIsMutual)




    //message := "YOLO\x00\x00\x00\x00\x00\x00\x00\x92"
    conn.Write(frame)

    //n, err := io.WriteString(conn, frame)
    /*if err != nil {
        log.Fatalf("client: write: %s", err)
    }*/
    //log.Printf("client: wrote %q (%d bytes)", message, n)
/*
    reply := make([]byte, 256)
    n, err = conn.Read(reply)
    log.Printf("client: read %q (%d bytes)", string(reply[:n]), n)
    log.Print("client: exiting")*/
}
