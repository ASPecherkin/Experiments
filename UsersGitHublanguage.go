package main

import (
	"fmt"
	"sync"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"os"
)

// LangValue struct for github.Repository.ListLanguages map
type LangValue struct {
	Lang string
	Value int
}

// RepoLang struct for identify repo with his used languages
type RepoLang struct {
	RepName string
    langs LangValue
}

// Stats for concurrency accessible slice from goroutines
type Stats struct {
	sync.RWMutex
	Statistic []RepoLang
}

func (s *Stats) add(name, login string, client *github.Client, wg *sync.WaitGroup, ) {
    defer s.Unlock()
    defer wg.Done()
	langs,_,err := client.Repositories.ListLanguages(login,name)
	if err != nil {
		fmt.Println(err)
	}
	s.Lock()
    for k,v := range langs {
		tlang := LangValue{Lang:k, Value: v}
		s.Statistic = append(s.Statistic, RepoLang{RepName:name, langs:tlang})
	}
}

func main() {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken:"some token here"})
	tc := oauth2.NewClient(oauth2.NoContext,ts)
	client := github.NewClient(tc)
	result := Stats{Statistic:make([]RepoLang,0,0)}
	var wg sync.WaitGroup
	repos, _, err := client.Repositories.List(os.Args[1], nil)
	if err != nil {
		fmt.Println(err)
	}
	for _, v := range repos {
		fmt.Println(*v.Name)
		wg.Add(1)
		go result.add(*v.Name,"envek",client,&wg)
	}
	wg.Wait()
	fmt.Println(result)
}
