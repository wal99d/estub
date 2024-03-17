package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/gofiber/fiber/v2"
)

const (
	MAX_FILE_SIZE = 1
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

type fileWriter struct {
	bytesWritten int
}

func (fw *fileWriter) Write(b []byte) (int, error) {
	fw.bytesWritten = len(b)
	if fw.bytesWritten > MAX_FILE_SIZE {
		return 0, fmt.Errorf("max file size (%d) exceeded", MAX_FILE_SIZE)
	}
	return fw.bytesWritten, nil
}

func main() {
	app := fiber.New(fiber.Config{
		ReadTimeout: 5 * time.Second,
	})
	resp := new(bytes.Buffer)
	fw := fileWriter{}
	done := make(chan struct{})
	sig := make(chan os.Signal)
	uri := generateURI()
	baseUrl := "http://localhost:3000"

	ssh.Handle(func(sess ssh.Session) {
		tee := io.TeeReader(sess, &fw)
		_, err := resp.ReadFrom(tee)
		if err != nil {
			close(done)
			sess.Write([]byte(fmt.Sprintf("error: %s", err)))
		} else {
			sess.Write([]byte("0x0C"))
			url := fmt.Sprintf("%s/%s", baseUrl, uri)
			sess.Write([]byte(url))
		}
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
