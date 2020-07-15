package db

import (
	"database/sql"
	"log"
)

type LinksTable struct {
	db   *sql.DB
	fast map[string]string
}

func GetLinksTable(db *sql.DB) LinksTable {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS links (discord_id TEXT UNIQUE NOT NULL, player_id TEXT UNIQUE NOT NULL)")

	if err != nil {
		log.Fatalln("Failed to init authentication table\n" + err.Error())
	}
	return LinksTable{
		db:   db,
		fast: map[string]string{},
	}
}

func (lt *LinksTable) SetLink(discordID string, playerID string) error {
	// check if they already have a link
	oldID := lt.GetPlayerID(discordID)

	if len(oldID) > 0 {
		prep, _ := lt.db.Prepare(
			"UPDATE links SET discord_id=? AND player_id=? WHERE discord_id=? OR player_id=?",
		)

		_, err := prep.Exec(discordID, playerID, discordID, playerID)

		if err != nil {
			log.Printf(
				"Failed to set (%s/%s), because\n%s\n",
				discordID, playerID, err.Error(),
			)
		} else {
			go lt.fastStore(playerID, discordID)
		}

		return err
	} else {
		return lt.NewLink(discordID, playerID)
	}
}

func (lt *LinksTable) NewLink(discordID string, playerID string) error {
	prep, _ := lt.db.Prepare("INSERT INTO links (discord_id, player_id) VALUES (?,?)")
	_, err := prep.Exec(discordID, playerID)

	if err != nil {
		log.Printf(
			"Failed to insert (%s/%s), because\n%s\n",
			discordID, playerID, err.Error(),
		)
	} else {
		go lt.fastStore(playerID, discordID)
	}

	return err
}

func (lt *LinksTable) UnLink(identifier string) error {
	prep, _ := lt.db.Prepare("DELETE FROM links WHERE discord_id=? OR player_id=?")
	_, err := prep.Exec(identifier, identifier)

	if err != nil {
		log.Printf(
			"Failed to remove (%s), because\n%s\n",
			identifier, err.Error(),
		)
	} else {
		lt.fastRemove(identifier)
	}

	return err
}

func (lt *LinksTable) GetPlayerID(discordID string) (playerID string) {
	playerID, isOK := lt.fastLoad(discordID)

	if isOK {
		return playerID
	}

	prep, _ := lt.db.Prepare(
		"SELECT player_id FROM links WHERE discord_id=?",
	)

	rows, err := prep.Query(discordID)

	if err != nil {
		log.Printf("Failed to get \"%s\"'s player ID, because\n%s\n", discordID, err.Error())
		return ""
	}

	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&playerID)

		if err != nil {
			log.Printf(
				"Failed to get \"%s\"'s player ID, because\n%s\n",
				discordID,
				err.Error(),
			)
			return ""
		} else {
			go lt.fastStore(playerID, discordID)
			return playerID
		}
	}
	return ""
}

func (lt *LinksTable) GetDiscordID(playerID string) (discordID string) {
	discordID, isOK := lt.fastLoad(playerID)

	if isOK {
		return discordID
	}

	prep, _ := lt.db.Prepare(
		"SELECT discord_id FROM links WHERE player_id=?",
	)

	rows, err := prep.Query(playerID)

	if err != nil {
		log.Printf("Failed to get \"%s\"'s player ID, because\n%s\n", playerID, err.Error())
		return ""
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&discordID)

		if err != nil {
			log.Printf(
				"Failed to get \"%s\"'s player ID, because\n%s\n",
				playerID,
				err.Error(),
			)
			return ""
		} else {
			go lt.fastStore(discordID, playerID)
			return discordID
		}
	}
	return ""
}

func (lt *LinksTable) fastStore(playerID string, discordID string) {
	lt.fast[playerID] = discordID
	lt.fast[discordID] = playerID
}

func (lt *LinksTable) fastRemove(identifier string) {
	discordID, isOK := lt.fast[identifier]
	if isOK {
		delete(lt.fast, discordID)
		delete(lt.fast, identifier)
	} else {
		playerID, isOK := lt.fast[identifier]
		if isOK {
			delete(lt.fast, identifier)
			delete(lt.fast, playerID)
		}
	}
}

func (lt *LinksTable) fastLoad(identifier string) (string, bool) {
	result, isOK := lt.fast[identifier]
	return result, isOK
}
