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

// Stats for concurrency accessible slice from goroutines
type Stats struct {
	sync.RWMutex
	Statistic map[string][]LangValue
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
	    s.Statistic[name] = append(s.Statistic[name], LangValue{Lang:k, Value: v})
	}
}

func main() {
	login := os.Args[1]
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken:"..."})
	tc := oauth2.NewClient(oauth2.NoContext,ts)
	client := github.NewClient(tc)
	result := Stats{Statistic:make(map[string][]LangValue)}
	var wg sync.WaitGroup
	repos, _, err := client.Repositories.List(login, nil)
	if err != nil {
		fmt.Println(err)
	}
	for _, v := range repos {
	   wg.Add(1)
	   go result.add(*v.Name,login,client,&wg)
	}
	wg.Wait()
	for k,v := range result.Statistic {
		fmt.Println(k)
		for _,pers := range v {
			fmt.Printf("%v %v \n", pers.Lang,pers.Value)
		}
		fmt.Println("#####")
	}
}
