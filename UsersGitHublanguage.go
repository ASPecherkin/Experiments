package main

import (
	"fmt"

	"sync"

	"github.com/google/go-github/github"
)

type Langs struct {
	sync.RWMutex
	m map[string]map[string]int
}

func (l *Langs) add(name, login string, client *github.Client, wg *sync.WaitGroup) {
	defer wg.Done()
	l.Lock()
	l.m[name], _, _ = client.Repositories.ListLanguages(login, name)
	fmt.Println(client.Repositories.ListLanguages(login, name))
	l.Unlock()
}

func main() {
	client := github.NewClient(nil)
	result := Langs{}
	var wg sync.WaitGroup

	repos, _, err := client.Repositories.List("envek", nil)
	if err != nil {
		fmt.Println(err)
	}
	for _, v := range repos {
		fmt.Println(*v.Name)
	}
	wg.Add(1)
	go result.add("addressable", "envek", client, &wg)
	if err != nil {
		fmt.Println(err)
	}
	wg.Wait()
	fmt.Println(result)

}
