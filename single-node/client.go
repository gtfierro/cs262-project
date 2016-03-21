package main

import (
	"net"
)

type Client struct {
	conn  *net.Conn
	query string
}

func NewClient(query string, conn *net.Conn) *Client {
	return &Client{query: query, conn: conn}
}
