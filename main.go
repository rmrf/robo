package main

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/common/log"
	"github.com/tj/docopt"
	"github.com/tj/robo/cli"
	"github.com/tj/robo/config"
)

var version = "0.5.5"

const usage = `
  Usage:
    robo [--config file]
    robo <task> [<arg>...] [--config file]
    robo help [<task>] [--config file]
    robo variables [--config file]
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

	// ---------------------

	r := gin.New()

	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// your custom format
		return fmt.Sprintf("%s - [%s] \"%s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))
	r.Use(gin.Recovery())
	r.GET("/:action/:userid", func(gc *gin.Context) {
		token := gc.DefaultQuery("token", "")
		if token != "sayMyN0me" {
			log.Warnf("Wrong token: %s", token)
			gc.JSON(http.StatusForbidden, gin.H{"message": "bad token"})
		} else {
			action := gc.Param("action")
			userid := gc.Param("userid")
			log.Infof("Doing %s: %s", action, userid)
			cli.Run(c, action, []string{userid})
			gc.JSON(http.StatusOK, gin.H{
				"message": "ok",
			})
		}
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
	// ---------------------

	switch {
	case args["help"].(bool):
		if name, ok := args["<task>"].(string); ok {
			cli.Help(c, name)
		} else {
			cli.List(c)
		}
	case args["variables"].(bool):
		cli.ListVariables(c)
	default:
		if name, ok := args["<task>"].(string); ok {
			cli.Run(c, name, args["<arg>"].([]string))
		} else {
			cli.List(c)
		}
	}
}
