// TODO handle errors at db level
package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"

	"github.com/zekurio/daemon/internal/models"
	"github.com/zekurio/daemon/internal/services/database"
	"github.com/zekurio/daemon/internal/services/database/dberr"
	"github.com/zekurio/daemon/internal/util/embedded"
	"github.com/zekurio/daemon/internal/util/vote"
	"github.com/zekurio/daemon/pkg/perms"
)

type Postgres struct {
	db *sql.DB
}

var (
	_           database.Database = (*Postgres)(nil)
	guildTables                   = []string{"guilds", "permissions"}
)

func InitPostgres(c models.PostgresConfig) (*Postgres, error) {
	var (
		p   Postgres
		err error
	)

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", c.Host, c.Port, c.Username, c.Password, c.Database)
	p.db, err = sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	err = p.db.Ping()
	if err != nil {
		return nil, err
	}

	err = p.handleMigrations()
	if err != nil {
		return nil, err
	}

	goose.SetBaseFS(embedded.Migrations)
	goose.SetDialect("postgres")
	goose.SetLogger(log.StandardLog())
	err = goose.Up(p.db, "migrations")
	if err != nil {
		return nil, err
	}

	return &p, nil

}

func (p *Postgres) handleMigrations() error {

	return nil

}

func (p *Postgres) Close() error {
	return p.db.Close()

}

// GUILDS

func (p *Postgres) GetAutoRoles(guildID string) ([]string, error) {
	roleStr, err := GetValue[string](p, "guilds", "autorole_ids", "guild_id", guildID)
	if roleStr == "" {
		return []string{}, err
	}

	return strings.Split(roleStr, ","), nil
}

func (p *Postgres) SetAutoRoles(guildID string, roleIDs []string) error {
	return SetValue(p, "guilds", "autorole_ids", strings.Join(roleIDs, ","), "guild_id", guildID)
}

func (p *Postgres) GetAutoVoice(guildID string) ([]string, error) {
	chStr, err := GetValue[string](p, "guilds", "autovoice_ids", "guild_id", guildID)
	if chStr == "" {
		return []string{}, err
	}

	return strings.Split(chStr, ","), nil
}

func (p *Postgres) SetAutoVoice(guildID string, channelIDs []string) error {
	return SetValue(p, "guilds", "autovoice_ids", strings.Join(channelIDs, ","), "guild_id", guildID)
}

func (p *Postgres) GetCreatedAV(guildID string) ([]string, error) {
	chStr, err := GetValue[string](p, "guilds", "created_av_ids", "guild_id", guildID)
	if chStr == "" {
		return []string{}, err
	}

	return strings.Split(chStr, ","), nil
}

func (p *Postgres) SetCreatedAV(guildID string, channelIDs []string) error {
	return SetValue(p, "guilds", "created_av_ids", strings.Join(channelIDs, ","), "guild_id", guildID)
}

// PERMISSIONS

func (p *Postgres) GetPermissions(guildID string) (map[string]perms.PermsArray, error) {
	results := make(map[string]perms.PermsArray)
	rows, err := p.db.Query(`SELECT role_id, perms FROM permissions WHERE guild_id = $1`, guildID)
	if err != nil {
		return nil, p.wrapErr(err)
	}

	for rows.Next() {
		var roleID string
		var permStr string

		err := rows.Scan(&roleID, &permStr)
		if err != nil {
			return nil, p.wrapErr(err)
		}

		results[roleID] = strings.Split(permStr, ",")
	}

	return results, nil
}

func (p *Postgres) SetPermissions(guildID, roleID string, perms perms.PermsArray) error {

	if len(perms) == 0 {
		_, err := p.db.Exec(`DELETE FROM permissions WHERE guild_id = $1 AND role_id = $2`, guildID, roleID)
		return err
	}

	pStr := strings.Join(perms, ",")
	res, err := p.db.Exec(`UPDATE permissions SET perms = $1 WHERE guild_id = $2 AND role_id = $3`, pStr, guildID, roleID)
	if err != nil {
		return err
	}
	ar, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if ar == 0 {
		_, err := p.db.Exec(`INSERT INTO permissions (guild_id, role_id, perms) VALUES ($1, $2, $3)`, guildID, roleID, pStr)
		return err
	}

	return nil

}

// ROLESELECTION

func (p *Postgres) AddRoleSelections(v []models.RoleSelection) error {

	var (
		err          error
		failedInsert = []models.RoleSelection{}
	)

	for _, r := range v {
		_, err := p.db.Exec(`INSERT INTO roleselection (guild_id, channel_id, message_id, role_id) VALUES ($1, $2, $3, $4)`, r.GuildID, r.ChannelID, r.MessageID, r.RoleID)
		if err != nil {
			failedInsert = append(failedInsert, r)
		}
	}

	if len(failedInsert) > 0 || err != nil {
		return fmt.Errorf("failed to insert roleselections: %v", failedInsert)
	}

	return nil

}

func (p *Postgres) GetRoleSelections() ([]models.RoleSelection, error) {

	var (
		err          error
		failedSelect = []models.RoleSelection{}
	)

	rows, err := p.db.Query(`SELECT guild_id, channel_id, message_id, role_id FROM roleselection`)
	if err != nil {
		return nil, err
	}

	var results []models.RoleSelection
	for rows.Next() {
		var r models.RoleSelection
		err := rows.Scan(&r.GuildID, &r.ChannelID, &r.MessageID, &r.RoleID)
		if err != nil {
			failedSelect = append(failedSelect, r)
		}
		results = append(results, r)
	}

	if len(failedSelect) > 0 || err != nil {
		return nil, fmt.Errorf("failed to select roleselections: %v", failedSelect)
	}

	return results, nil

}

func (p *Postgres) RemoveRoleSelections(guildID, channelID, messageID string) error {
	_, err := p.db.Exec(`DELETE FROM roleselection WHERE guild_id = $1 AND channel_id = $2 AND message_id = $3`, guildID, channelID, messageID)
	return err
}

// VOTES

func (p *Postgres) GetVotes() (map[string]vote.Vote, error) {

	rows, err := p.db.Query(`SELECT json_data, id FROM votes`)
	if err != nil {
		return nil, err
	}

	var results = make(map[string]vote.Vote)
	for rows.Next() {
		var voteID, rawData string
		err := rows.Scan(&voteID, &rawData)
		if err != nil {
			continue
		}
		vote, err := vote.Unmarshal(rawData)
		if err != nil {
			p.DeleteVote(rawData)
		} else {
			results[vote.ID] = vote
		}

	}

	return results, nil

}

func (p *Postgres) AddUpdateVote(v vote.Vote) error {
	rawData, err := vote.Marshal(v)
	if err != nil {
		return err
	}
	_, err = p.db.Exec(`INSERT INTO votes (json_data, id) VALUES ($1, $2) ON CONFLICT (id) DO UPDATE SET json_data = $1`, rawData, v.ID)
	return err
}

func (p *Postgres) DeleteVote(voteID string) error {
	_, err := p.db.Exec(`DELETE FROM votes WHERE id = $1`, voteID)
	return err
}

// DATA MANAGEMENT

func (p *Postgres) FlushGuildData(guildID string) error {

	return p.tx(func(tx *sql.Tx) error {

		var (
			err          error
			failedGuilds = []string{}
		)

		for _, table := range guildTables {

			_, err := tx.Exec(fmt.Sprintf(`DELETE FROM %s WHERE guild_id = $1`, table), guildID)
			if err != nil {
				failedGuilds = append(failedGuilds, guildID)
			}

		}

		if len(failedGuilds) > 0 || err != nil {
			return fmt.Errorf("failed to flush guild data for guilds: %v", failedGuilds)
		}

		return nil

	})

}

//
// HELPERS
//

func (p *Postgres) tx(f func(*sql.Tx) error) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}

	if err = f(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (p *Postgres) wrapErr(err error) error {
	if err != nil && err == sql.ErrNoRows {
		return dberr.ErrNotFound
	}
	return err
}

// GetValue retrieves a specific value from a PostgreSQL table.
func GetValue[TVal, TWv any](t *Postgres, table, valueKey, whereKey string, whereValue TWv) (TVal, error) {
	var value TVal
	query := fmt.Sprintf(`SELECT "%s" FROM %s WHERE "%s" = $1`, valueKey, table, whereKey)
	err := t.db.QueryRow(query, whereValue).Scan(&value)
	return value, t.wrapErr(err)
}

// SetValue updates a specific value in a PostgreSQL table, or inserts a new row if none is found.
func SetValue[TVal, TWv any](t *Postgres, table, valueKey string, value TVal, whereKey string, whereValue TWv) error {
	updateQuery := fmt.Sprintf(`UPDATE %s SET "%s" = $1 WHERE "%s" = $2`, table, valueKey, whereKey)
	result, err := t.db.Exec(updateQuery, value, whereValue)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		insertQuery := fmt.Sprintf(`INSERT INTO %s ("%s", "%s") VALUES ($1, $2)`, table, whereKey, valueKey)
		_, err = t.db.Exec(insertQuery, whereValue, value)
		if err == nil {
			return err
		}
	}

	return err
}
