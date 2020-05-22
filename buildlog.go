package main

const path = "data.db"
const bucket = "build"

type Buildlog struct {
	Name 				string
	Hash 				map[string]string
	BuildNumber string
	Image 			string
	Tag 				string
	UpdatedAt 	string
}

var db Db = Blot{path, bucket}

func (log Buildlog) SaveBuild() (Buildlog, error) {
	return db.save(log)
}

func (log Buildlog) DelBuild() error {
	return db.del(log.Name)
}

func ListBuild() ([]Buildlog, error) {
	return db.list()
}

func GetBuild(name string) (Buildlog, error) {
	return db.get(name)
}