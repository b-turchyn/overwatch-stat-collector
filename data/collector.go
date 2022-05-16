package data

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Player struct {
  Name string
  Number int
}

type PlayerStats struct {
  CollectionDate time.Time
  Player Player
  GamesPlayed int
  GamesWon int
  TankLevel int
  DamageLevel int
  SupportLevel int
}

type ApiResult struct {
  CompetitiveStats ApiCompStats `json:"competitiveStats"`
  Ratings []ApiRatings
}

type ApiCompStats struct {
  Games ApiGameStats `json:"games"`
}

type ApiGameStats struct {
  Played int `json:"played"`
  Won int `json:"won"`
}

type ApiRatings struct {
  Level int `json:"level"`
  Role string `json:"role"`
}

func CollectAll(players []Player) (map[Player]PlayerStats, error) {
  result := make(map[Player]PlayerStats)
  var err error

  for _, v := range players {
    time.Sleep(500 * time.Millisecond)
    stats, innerError := CollectUser(v)

    if innerError != nil {
      err = innerError
      break
    }
    result[v] = stats
  }

  return result, err
}

func CollectUser(player Player) (PlayerStats, error) {
  result := PlayerStats{
    Player: player,
    CollectionDate: time.Now().Local(),
  }

  result.Player = player
  resp, err := http.Get(player.url())

  if err != nil {
    return result, nil
  }
  defer resp.Body.Close()

  body, err := io.ReadAll(resp.Body)

  var apiData ApiResult
  err = json.Unmarshal(body, &apiData)
  if err != nil {
    return result, nil
  }

  result.GamesPlayed = apiData.CompetitiveStats.Games.Played
  result.GamesWon = apiData.CompetitiveStats.Games.Won

  for _, v := range apiData.Ratings {
    switch v.Role {
    case "tank":
      result.TankLevel = v.Level
    case "support":
      result.SupportLevel = v.Level
    case "damage":
      result.DamageLevel = v.Level
    }
  }

  return result, err
}

func (p Player) url() string {
  return fmt.Sprintf("https://ow-api.com/v1/stats/pc/us/%s-%d/profile", p.Name, p.Number)
}
