package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis"

	"../models"
	"../services"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

// SessionController contains receiver functions
// for file-hydra style file downloads.
type SessionController struct {
	// TODO: add redis session
	basePath string
	rc       *redis.Client
}

// NewSessionController configures a new session
// controller.
// Returns a pointer to the configured controller.
func NewSessionController(
	basePath string,
	rc *redis.Client,
) *SessionController {
	return &SessionController{
		basePath: basePath,
		rc:       rc,
	}
}

// CreateSession is an endpoint that takes a valid
// filename from the body and validates its existance
// in /public.
// Returns a UUID to identify the created session.
func (sc SessionController) CreateSession(
	w http.ResponseWriter,
	req *http.Request,
	_ httprouter.Params,
) {
	i := models.Init{}
	json.NewDecoder(req.Body).Decode(&i)

	if !isValidFilename(i.Filename) {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Must provide valid filename.")
		return
	}

	if i.ExpectedClients < 2 {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Must expect at least two clients.")
		return
	}

	p := sc.createPath(i.Filename)

	if !services.FileExists(p) {
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, "File does not exist.")
		return
	}

	sess := models.Session{
		ID:               uuid.New().String(),
		Filename:         i.Filename,
		ConnectedClients: 0,
		ExpectedClients:  i.ExpectedClients,
	}

	sessjson, _ := json.Marshal(sess)

	err := sc.rc.Set(req.Context(), sess.ID, sessjson, time.Minute*5).Err()

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, string(sessjson))
}

// GetSession requires an id to be sent as a path variable
// and returns the given session for that id or 404 not found.
func (sc SessionController) GetSession(
	w http.ResponseWriter,
	req *http.Request,
	p httprouter.Params,
) {
	id := p.ByName("id")

	sessjson, err := sc.rc.Get(req.Context(), id).Result()

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var sess models.Session
	json.Unmarshal([]byte(sessjson), &sess)

	if sess.ConnectedClients >= sess.ExpectedClients {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "All expected clients already connected.")
		return
	}

	sid := sess.ConnectedClients
	sess.ConnectedClients++

	buf, _ := json.Marshal(sess)
	sessjson = string(buf)

	sc.rc.Set(req.Context(), id, sessjson, time.Minute*5)
	sc.rc.Publish(req.Context(), id, sessjson)

	awaitAllClients(req.Context(), sess, sc.rc)

	f, err := os.Open(sc.createPath(sess.Filename))
	defer f.Close()

	if err != nil {
		log.Fatalln(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fs, _ := f.Stat()
	off, end := getOffsetAndEnd(
		fs.Size(),
		int64(sid),
		int64(sess.ExpectedClients),
	)
	f.Seek(off, 0)
	fbuf := make([]byte, end-off)
	f.Read(fbuf)
	w.Write(fbuf)
}

func awaitAllClients(ctx context.Context, sess models.Session, c *redis.Client) {
	sub := c.Subscribe(ctx, sess.ID)
	_, err := sub.Receive(ctx)

	if err != nil {
		log.Println(err)
	}

	diff := sess.ExpectedClients - sess.ConnectedClients
	for i := 0; i < diff; i++ {
		msg, _ := sub.ReceiveMessage(ctx)
		handleMessage(msg)
	}
}

func handleMessage(msg *redis.Message) {
	var sess models.Session
	json.Unmarshal([]byte(msg.Payload), &sess)

	if sess.ExpectedClients == sess.ConnectedClients {
		log.Println("All clients connected!")
	}
}

func getOffsetAndEnd(size int64, sid int64, ec int64) (int64, int64) {
	len := size / ec
	off := len * sid
	var end int64
	if sid == ec-1 {
		end = off + len + size%ec
	} else {
		end = off + len
	}
	return off, end
}

func (sc SessionController) createPath(fn string) string {
	return fmt.Sprintf("%s/%s", sc.basePath, fn)
}

func isValidFilename(fn string) bool {
	fn = strings.TrimSpace(fn)
	return fn != "" && !strings.Contains(fn, "../")
}
