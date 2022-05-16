package db

import (
	"database/sql"

	"github.com/b-turchyn/overwatch-stat-collector/data"
	"github.com/b-turchyn/overwatch-stat-collector/util"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

func OpenDatabase(dbname string) (*sql.DB, error) {
  db, err := sql.Open("sqlite3", dbname)

  if err == nil {
    err = createUsersTable(db)
  }
  if err == nil {
    err = createStatsTable(db)
  }

  return db, err
}

func createUsersTable(db *sql.DB) error {
  // Table structure is based on the BattleTag Naming Policy, with some wiggle room on the name size
  // https://us.battle.net/support/en/article/26963
  stmt, err := db.Prepare(`
  CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY,
    name VARCHAR(16) NOT NULL,
    number INTEGER NOT NULL
  )
  `)

  if err != nil {
    return err
  }
  defer stmt.Close()

  _, err = stmt.Exec()

  return err
}

func createStatsTable(db *sql.DB) error {
  stmt, err := db.Prepare(`
  CREATE TABLE stats (
    id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL,
    created_at DATETIME NOT NULL,
    games_played INTEGER NOT NULL,
    games_won INTEGER NOT NULL,
    tank_level INTEGER,
    damage_level INTEGER,
    support_level INTEGER
  )
  `)

  if err != nil {
    return err
  }
  defer stmt.Close()

  _, err = stmt.Exec()

  return err
}

func GetAllUsers(db *sql.DB) ([]data.Player, error) {
  util.Logger.Debug("Selecting all users")
  var result []data.Player
  stmt, err := db.Prepare("SELECT name, number FROM users ORDER BY id")

  if err != nil {
    return result, err
  }
  defer stmt.Close()

  rows, err := stmt.Query()
  defer rows.Close()

  if err != nil {
    return result, err
  }

  for rows.Next() {
    var row data.Player

    rows.Scan(&row.Name, &row.Number)
    result = append(result, row)
  }

  return result, err
}

func InsertAllPlayerStats(db *sql.DB, playerStats map[data.Player]data.PlayerStats) error {
  util.Logger.Debug("Inserting all stats")
  for _, v := range playerStats {
    err := InsertPlayerStats(db, v)
    if err != nil {
      return err
    }
  }

  return nil
}

func GetPlayerStats(db *sql.DB, p data.Player) ([]data.PlayerStats, error) {
  var result []data.PlayerStats

  stmt, err := db.Prepare(`
    SELECT created_at, games_played, games_won, tank_level, damage_level, support_level
    FROM stats s
    INNER JOIN users u ON u.id = s.user_id
    AND u.name = ? AND u.number = ?
    ORDER BY s.created_at
  `)

  if err != nil {
    return nil, err
  }
  defer stmt.Close()

  rows, err := stmt.Query(p.Name, p.Number)

  for rows.Next() {
    row := data.PlayerStats{
      Player: p,
    }

    rows.Scan(&row.CollectionDate, &row.GamesPlayed, &row.GamesWon, &row.TankLevel, &row.DamageLevel, &row.SupportLevel)

    result = append(result, row)
  }

  return result, nil
}

func InsertPlayerStats(db *sql.DB, p data.PlayerStats) error {
  util.Logger.Debug("Inserting stats", zap.String("name", p.Player.Name), zap.Int("number", p.Player.Number))
  stmt, err := db.Prepare(`INSERT INTO stats
    (user_id, created_at, games_played, games_won, tank_level, damage_level, support_level)
    VALUES
    ((SELECT id FROM users WHERE name = ? AND number = ?), ?, ?, ?, ?, ?, ?)`)

  if err != nil {
    return err
  }
  defer stmt.Close()

  _, err = stmt.Exec(
    p.Player.Name,
    p.Player.Number,
    p.CollectionDate,
    p.GamesPlayed,
    p.GamesWon,
    p.TankLevel,
    p.DamageLevel,
    p.SupportLevel,
  )

  return err
}
