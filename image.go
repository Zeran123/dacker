package main

type Image struct {
	Name 				string
	Dockerfile 	string 
	Image 			string
	Tag 				string
	Deps				[]string
}