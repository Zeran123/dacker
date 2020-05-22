package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"errors"

	bolt "go.etcd.io/bbolt"
)

type Db interface {
	save(build Buildlog) (Buildlog, error)
	list() ([]Buildlog, error)
	get(id string) (Buildlog, error)
	del(id string) error
}

type Blot struct {
	path   string
	bucket string
}

func (b Blot) save(build Buildlog) (Buildlog, error) {
	db, err := bolt.Open(b.path, 0600, nil)
	if err != nil {
		return build, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b, _ := tx.CreateBucketIfNotExists([]byte(b.bucket))
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		err := enc.Encode(build)
		err = b.Put([]byte(build.Name), buf.Bytes())
		return err
	})
	defer db.Close()
	if err != nil {
		return build, err
	}
	return build, nil
}

func (b Blot) list() ([]Buildlog, error) {
	builds := make([]Buildlog, 0, 5)
	db, err := bolt.Open(b.path, 0600, nil)
	if err != nil {
		return nil, err
	}
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(b.bucket))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var build Buildlog
			buf := bytes.NewBuffer(v)
			enc := gob.NewDecoder(buf)
			err := enc.Decode(&build)
			if err != nil {
				log.Fatal(err)
			}
			builds = append(builds, build)
		}
		return err
	})
	defer db.Close()
	if err != nil {
		return nil, err
	}
	return builds, nil
}

func (b Blot) get(name string) (Buildlog, error) {
	var build Buildlog
	db, err := bolt.Open(b.path, 0600, nil)
	if err != nil {
		return build, err
	}
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(b.bucket))
		if b == nil {
			return errors.New("bucket is empty")
		}
		v := b.Get([]byte(name))
		buf := bytes.NewBuffer(v)
		enc := gob.NewDecoder(buf)
		err := enc.Decode(&build)
		return err
	})
	defer db.Close()
	if err != nil {
		return build, err
	}
	return build, nil
}

func (b Blot) del(name string) error {
	db, err := bolt.Open(b.path, 0600, nil)
	if err != nil {
		return err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(b.bucket))
		err := b.Delete([]byte(name))
		return err
	})
	defer db.Close()
	return nil
}
