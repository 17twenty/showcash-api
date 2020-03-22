package showcash

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
)

// DAO is the Data Abstraction Object we use to seperate out the database
type DAO struct {
	db *sqlx.DB
}

// NewDAO is a postgres backed version of the DAO
func NewDAO(username, password string, host string, port int, database string) (*DAO, error) {
	uri := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		username,
		password,
		host,
		port,
		database,
	)
	d := &DAO{}

	var err error
	d.db, err = sqlx.Open("postgres", uri)
	if err != nil {
		log.Println("Couldn't Open() DB with:",
			username,
			"******",
			host,
			port,
			database,
			err)
		return nil, errBadDAO
	}

	if !d.IsConnected() {
		log.Println("Couldn't Ping() DB with:",
			username,
			"******",
			host,
			port,
			database,
			err)
		return nil, errBadDAO
	}
	// Stops us having to add the db tags
	d.db.Mapper = reflectx.NewMapperFunc("json", strings.ToLower)
	return d, nil
}

// IsConnected is a healthcheck for the DAO
// In memory versions would be static but DB backed would ping()
func (d *DAO) IsConnected() bool {
	if d == nil {
		return false
	}
	if err := d.db.Ping(); err != nil {
		log.Println("Error from database ", err)
		return false
	}
	return true
}

// createPost WONT insert items from the item list by default
func (d *DAO) createPost(userID uuid.UUID, p Post) (Post, error) {
	p.ID = uuid.Must(uuid.NewV4())
	p.Date = time.Now()

	_, err := d.db.NamedExec(
		`INSERT INTO showcash.post(
			id,
			title,
			imageuri,
			date
		) VALUES (
			:id,
			:title,
			:imageuri,
			:date
		)`, &p)
	if err != nil {
		return Post{}, err
	}

	// Note: need for items here
	return p, err
}

func (d *DAO) updatePost(userID uuid.UUID, p Post) (Post, error) {
	res, err := d.db.Exec(
		`UPDATE showcash.post SET
			title = $1,
			imageuri = $2
			WHERE user_id = $3 AND id = $4`,
		p.Title,
		p.ImageURI,
		userID,
		p.ID,
	)
	if err != nil {
		return Post{}, err
	}
	if cnt, err := res.RowsAffected(); err != nil || cnt == 0 {
		return Post{}, sql.ErrNoRows
	}

	for i := range p.ItemList {
		_, err = d.db.Exec(
			`INSERT INTO showcash.item (
				post_id,
				id,
				title,
				description,
				link,
				"left",
				top
			) VALUES (
				 $1, $2, $3, $4, $5, $6, $7
			) 
			ON CONFLICT (post_id, id) 
			DO UPDATE SET
				title = EXCLUDED.title,
				description = EXCLUDED.description,
				link = EXCLUDED.link,
				"left" = EXCLUDED."left",
				top = EXCLUDED.top
				`,
			p.ID,
			p.ItemList[i].ID,
			p.ItemList[i].Title,
			p.ItemList[i].Description,
			p.ItemList[i].Link,
			p.ItemList[i].Left,
			p.ItemList[i].Top,
		)
		if err != nil {
			log.Println("Fucked UPSERT for post", p.ID, p.ItemList[i])
			break
		}
	}
	return p, err
}

func (d *DAO) getPost(postID uuid.UUID) (Post, error) {
	p := Post{}
	if err := d.db.Get(
		&p,
		`SELECT 
			id,
			title,
			imageuri,
			date
		FROM
			showcash.post
		WHERE id = $1
		LIMIT 1`, postID,
	); err != nil {
		return p, err
	}

	// Got it hydrate
	var items []Item
	err := d.db.Select(
		&items,
		`SELECT
			id,
			title,
			description,
			link,
			"left",
			top
		FROM 
			showcash.item
		WHERE post_id = $1`, p.ID)
	if len(items) > 0 {
		// Connect the Items to the Post
		p.ItemList = items
	}
	return p, err

}

func (d *DAO) getLatestPosts() {
}

func (d *DAO) getPopularPosts() {
}
