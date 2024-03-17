package main

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/gofiber/fiber/v2"
)

var (
	Characters = []rune("abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789")
)

func generateURI() string {
	uriString := make([]rune, 10)
	for i, uri := range uriString {
		uri = Characters[rand.Intn(len(uriString))]
		uriString[i] = uri
	}
	return string(uriString)
}

func init() {
	rand.New(rand.NewSource(time.Now().UnixMicro()))
}

func main() {
	app := fiber.New(fiber.Config{
		ReadTimeout: 5 * time.Second,
	})
	resp := new(bytes.Buffer)
	done := make(chan struct{})
	sig := make(chan os.Signal)
	uri := generateURI()
	baseUrl := "http://localhost:3000"

	ssh.Handle(func(sess ssh.Session) {
		resp.ReadFrom(sess)
		sess.Write([]byte(fmt.Sprintf("%x", "0x0C")))
		url := fmt.Sprintf("%s/%s", baseUrl, uri)
		sess.Write([]byte(url))
		<-done
		sess.Write([]byte("we are done"))
		sig <- os.Kill
	})

	app.Get("/:uri", func(c *fiber.Ctx) error {
		newUri := c.Params("uri")
		if strings.EqualFold(uri, newUri) {
			close(done)
			return c.SendStream(resp, resp.Len())
		} else {
			return c.SendString("Nothing to show..")
		}
	})
	go func() {
		log.Fatal(app.Listen(":3000"))
	}()

	go func(app *fiber.App) {
		<-sig
		os.Exit(0)
	}(app)
	defer app.Shutdown()

	publicKeyOpts := ssh.PublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
		data, _ := os.ReadFile("/Users/waleedalharthi/.ssh/id_rsa.pub")
		allowed, _, _, _, _ := ssh.ParseAuthorizedKey(data)
		return ssh.KeysEqual(key, allowed)
	})

	hostKeyFileOpts := ssh.HostKeyFile("/Users/waleedalharthi/.ssh/id_rsa")

	log.Println("starting ssh server on port 2222...")
	log.Fatal(ssh.ListenAndServe(":2222", nil, hostKeyFileOpts, publicKeyOpts))
}
