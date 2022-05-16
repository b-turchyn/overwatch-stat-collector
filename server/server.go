/*
Copyright © 2022 Brian Turchyn

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package server

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/b-turchyn/overwatch-stat-collector/data"
	"github.com/b-turchyn/overwatch-stat-collector/db"
	"github.com/b-turchyn/overwatch-stat-collector/util"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func InitServer(database *sql.DB) {
  r := gin.Default()

  var base gin.IRouter

  if viper.GetString("auth.username") != "" {
    base = initBasicAuth(r)
  } else {
    base = r
  }

  base.GET("/users", func(c *gin.Context) {
    players, err := db.GetAllUsers(database)

    if err != nil {
      c.AbortWithError(http.StatusInternalServerError, err)
    } else {
      c.JSON(http.StatusOK, players)
    }
  })

  base.GET("/users/:battletag", func(c *gin.Context) {
    util.Logger.Info(c.Param("battletag"))
    battletag := strings.Split(c.Param("battletag"), "-")

    if len(battletag) != 2 {
      c.JSON(http.StatusBadRequest, gin.H{"error": "Bad BattleTag format"})
    } else {
      number, err := strconv.Atoi(battletag[1])
      if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Bad BattleTag format"})
      } else {
        player := data.Player{
          Name: battletag[0],
          Number: number,
        }

        stats, err := db.GetPlayerStats(database, player)

        if err != nil {
          c.JSON(http.StatusInternalServerError, gin.H{"error": err})
        } else {
          c.JSON(http.StatusOK, stats)
        }
      }
    }
  })

  base.GET("/data.sqlite3", func(c *gin.Context) {
    c.File(viper.GetString("database.name"))
  })

  base.StaticFile("/", "./static/index.html")
  base.Static("/assets", "./static/assets")

  r.Run()
}

func initBasicAuth(r *gin.Engine) *gin.RouterGroup {
  util.Logger.Info("Enabling BasicAuth")
  result := r.Group("/", gin.BasicAuth(gin.Accounts{
    viper.GetString("auth.username"): viper.GetString("auth.password"),
  }))

  return result
}
