package main

const path = "data.db"
const bucket = "build"

type Buildlog struct {
	Name        string
	Hash        map[string]string
	BuildNumber string
	Image       string
	Tag         string
	Deps        map[string]string
	ReleaseTag  string
	ReleaseRef  string
	UpdatedAt   string
}

var db Db = Blot{path, bucket}

func (build Buildlog) SaveBuild() (Buildlog, error) {
	return db.save(build)
}

func (build Buildlog) DelBuild() error {
	return db.del(build.Name)
}

func ListBuild() ([]Buildlog, error) {
	return db.list()
}

func GetBuild(name string) (Buildlog, error) {
	return db.get(name)
}
