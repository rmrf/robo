package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/docopt/docopt-go"
	"github.com/gin-gonic/gin"
	"github.com/rmrf/robo/cli"
	"github.com/rmrf/robo/config"
)

var version = "0.5.6"

const usage = `
  Usage:
    robo [--config file]
    robo <task> [<arg>...] [--config file]
    robo help [<task>] [--config file]
    robo variables [--config file]
    robo startweb [--config file]
    robo -h | --help
    robo --version

  Options:
    -c, --config file   config file to load [default: robo.yml]
    -h, --help          output help information
    -v, --version       output version

  Examples:

    output tasks
    $ robo

    output task help
    $ robo help mytask

`

func main() {
	args, err := docopt.Parse(usage, nil, true, version, true)
	if err != nil {
		cli.Fatalf("error parsing arguments: %s", err)
	}

	abs, err := filepath.Abs(args["--config"].(string))
	if err != nil {
		cli.Fatalf("cannot resolve --config: %s", err)
	}

	c, err := config.New(abs)
	if err != nil {
		cli.Fatalf("error loading configuration: %s", err)
	}

	switch {
	case args["help"].(bool):
		if name, ok := args["<task>"].(string); ok {
			cli.Help(c, name)
		} else {
			cli.List(c)
		}
	case args["variables"].(bool):
		cli.ListVariables(c)
	case args["startweb"].(bool):
		roboV := c.Variables["robo"].(map[interface{}]interface{})
		startWeb(c, roboV["web-addr"].(string), roboV["token"].(string))
	default:
		if name, ok := args["<task>"].(string); ok {
			cli.Run(c, name, args["<arg>"].([]string))
		} else {
			cli.List(c)
		}
	}
}

func startWeb(conf *config.Config, addr, token string) {
	type PostBody struct {
		Token string   `json:"token""  binding:"required"`
		Args  []string `json:"args""  binding:"required"`
	}

	r := gin.New()
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// your custom format
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))
	r.Use(gin.Recovery())
	r.POST("/task/:taskname", func(gc *gin.Context) {
		var pBody PostBody
		if err := gc.ShouldBindJSON(&pBody); err != nil {
			gc.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if pBody.Token != token {
			log.Printf("Wrong token: %s", token)
			gc.JSON(http.StatusForbidden, gin.H{"message": "bad token"})
			return
		}
		taskName := gc.Param("taskname")
		info := fmt.Sprintf("%s: %s", taskName, pBody.Args)
		log.Println(info)
		err := cli.Run(conf, taskName, pBody.Args)
		if err != nil {
			log.Printf("client run failed: %s", err.Error())
			gc.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
		gc.JSON(http.StatusOK, gin.H{"message": info})
	})
	s := &http.Server{Addr: addr,
		Handler:      r,
		ReadTimeout:  6 * time.Second,
		WriteTimeout: 6 * time.Second}
	s.ListenAndServe()

}
