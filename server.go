package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

const addr = ":3333"
const backendService = "http://localhost:3333/back"

func main() {

	// I run backend service here because I'm lazy...
	// Assume this endpoint is on other server.
	http.HandleFunc("/back", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		return
	})

	// No timeout and no cancelation
	http.HandleFunc("/1", handler1)

	// Timeout but no cancelation
	http.HandleFunc("/2", handler2)

	// Timeout and canncel with done channel
	http.HandleFunc("/3", handler3)

	// Timeout with context
	http.HandleFunc("/4", handler4)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

func handler1(w http.ResponseWriter, r *http.Request) {
	errCh := make(chan error, 1)
	go func() {
		errCh <- request1()
	}()

	select {
	case err := <-errCh:
		if err != nil {
			log.Println("failed:", err)
			return
		}
	}

	log.Println("success")
}

func handler2(w http.ResponseWriter, r *http.Request) {
	errCh := make(chan error, 1)
	go func() {
		errCh <- request1()
	}()

	select {
	case err := <-errCh:
		if err != nil {
			log.Println("failed:", err)
			return
		}
	case <-time.After(2 * time.Second):
		log.Println("failed: timeout")
		return
	}

	log.Println("success")
}

func handler3(w http.ResponseWriter, r *http.Request) {

	doneCh := make(chan struct{}, 1)
	errCh := make(chan error, 1)
	go func() {
		errCh <- request2(doneCh)
	}()

	go func() {
		<-time.After(2 * time.Second)
		close(doneCh)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			log.Println("failed:", err)
			return
		}
	}

	log.Println("success")
}

func handler4(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- request3(ctx)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			log.Println("failed:", err)
			return
		}
	}

	log.Println("success")
}

func request1() error {
	log.Println("start request1")
	tr := &http.Transport{}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("POST", backendService, nil)
	if err != nil {
		return err
	}

	errCh := make(chan error, 1)
	go func() {
		_, err := client.Do(req)
		errCh <- err
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	}

	log.Println("end request1")
	return nil
}

func request2(doneCh chan struct{}) error {
	log.Println("start request2")
	tr := &http.Transport{}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("POST", backendService, nil)
	if err != nil {
		return err
	}

	errCh := make(chan error, 1)
	go func() {
		_, err := client.Do(req)
		errCh <- err
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case <-doneCh:
		tr.CancelRequest(req)
		<-errCh
		log.Println("cancel request2")
		return fmt.Errorf("canceled")
	}

	log.Println("end request2")
	return nil
}

func request3(ctx context.Context) error {
	log.Println("start request3")
	tr := &http.Transport{}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("POST", backendService, nil)
	if err != nil {
		return err
	}

	errCh := make(chan error, 1)
	go func() {
		_, err := client.Do(req)
		errCh <- err
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case <-ctx.Done():
		tr.CancelRequest(req)
		<-errCh
		log.Println("cancel request3")
		return ctx.Err()
	}

	log.Println("end request3")
	return nil
}
