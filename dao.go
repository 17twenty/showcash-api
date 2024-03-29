package showcash

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
	"github.com/lib/pq"
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

	_, err := d.db.Exec(
		`INSERT INTO showcash.post(
			user_id,
			id,
			title,
			imageuri,
			date
		) VALUES (
			$1, $2, $3, $4, $5
		)`, userID, p.ID, p.Title, p.ImageURI, p.Date)
	if err != nil {
		return Post{}, err
	}

	// Note: need for items here
	return p, err
}

func (d *DAO) claimPost(userID uuid.UUID, postID uuid.UUID) error {
	if postID == uuid.Nil {
		_, err := d.db.Exec(
			`UPDATE showcash.post SET 
				user_id = $1
			`, userID)
		return err
	}
	_, err := d.db.Exec(
		`UPDATE showcash.post SET 
			user_id = $1
			WHERE id = $2
		`, userID, postID)
	return err
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
			p.id,
			p.title,
			p.imageuri,
			p.date,
			u.username
		FROM
			showcash.post AS p JOIN showcash.user AS u ON u.user_id = p.user_id
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

func (d *DAO) deletePost(postID uuid.UUID) {
	_, err := d.db.Exec(
		`DELETE FROM showcash.post WHERE id = $1`,
		postID,
	)
	if err != nil {
		log.Println("Fucked DELETE for post", postID, err)
	}
	if _, err := d.db.Exec(
		`DELETE FROM showcash.item WHERE post_id = $1`,
		postID,
	); err != nil {
		log.Println("Fucked DELETE for items", postID, err)
	}
}

// increaseView keys on the postID and the unique value to ensure we're not being
// dickheads
func (d *DAO) increaseView(postID uuid.UUID, uniqueValue string) {
	_, err := d.db.Exec(
		`INSERT INTO showcash.views (
			post_id,
			unique_value
		) VALUES (
			$1, $2
		) ON CONFLICT (post_id, unique_value) DO NOTHING`,
		postID, uniqueValue,
	)
	if err != nil {
		log.Println("increaseView() Failed")
	}
}

func (d *DAO) getLatestPosts() []Post {
	var posts []Post
	err := d.db.Select(
		&posts,
		`SELECT p.id,p.imageuri,p.title,p.date,
		u.username FROM showcash.post AS p JOIN showcash.user AS u ON p.user_id = u.user_id
		ORDER BY p.created_at DESC   
		LIMIT 8`,
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println("getLatestPosts() failed", err)
	}
	return posts
}

func (d *DAO) getUsersLatestPosts(userID uuid.UUID) []Post {
	var posts []Post
	err := d.db.Select(
		&posts,
		`SELECT p.id,p.imageuri,p.title,p.date,
		u.username FROM showcash.post AS p JOIN showcash.user AS u ON p.user_id = u.user_id
		WHERE u.user_id = $1
		ORDER BY p.created_at DESC   
		LIMIT 50`, userID,
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println("getLatestPosts() failed", err)
	}
	return posts
}

func (d *DAO) getMostViewedPosts() []Post {
	var posts []Post
	// Porting from http://restfulmvc.com/reddit-algorithm.shtml
	err := d.db.Select(
		&posts,
		`SELECT p.id,p.imageuri,p.title,p.date,u.username FROM showcash.post AS p JOIN showcash.user AS u ON p.user_id = u.user_id JOIN LATERAL (  
			SELECT post_id, LOG(10,COUNT(*) + 1) * 287015 + ( SELECT extract(epoch FROM p.date)) AS rating   
			FROM showcash.views GROUP BY views.post_id 
		) AS pop ON pop.post_id = p.id ORDER BY pop.rating DESC LIMIT 8`,
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println("getMostViewedPosts() failed", err)
	}
	return posts
}

func (d *DAO) getCommentsForPostID(postID uuid.UUID) []Comment {
	var comments []Comment
	err := d.db.Select(
		&comments,
		`SELECT
			id,
			date,
			comment,
			username,
			user_id
		FROM 
			showcash.comments
		WHERE post_id = $1`, postID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println("getCommentsForPostID() failed", err)
	}
	return comments
}

// createComment WONT insert items from the item list by default
func (d *DAO) createComment(userID uuid.UUID, postID uuid.UUID, c Comment) (Comment, error) {
	c.ID = uuid.Must(uuid.NewV4())
	c.Date = time.Now()

	// TODO: Add userID/usernames etc
	_, err := d.db.Exec(
		`INSERT INTO showcash.comments(
			post_id,
			id,
			date,
			comment,
			username,
			user_id
		) VALUES (
			$1, $2, $3, $4, $5, $6
		)`, postID, c.ID, c.Date, c.Comment, c.Username, userID,
	)
	return c, err
}

func (d *DAO) getUserProfileByID(userID uuid.UUID) (UserProfile, error) {
	up := UserProfile{}
	err := d.db.Get(&up,
		`SELECT
			username,
			realname,
			location,
			profile_uri,
			bio,
			social_1,
			social_2,
			social_3,
			created_at
		FROM 
			showcash.user
		WHERE user_id = $1`, userID,
	)

	up.Interests = []string{"MFA"}
	up.Friends = []UserProfile{}
	up.Followers = []UserProfile{}

	return up, err
}

func (d *DAO) getUserProfileByHandle(handle string) (UserProfile, error) {
	up := UserProfile{}
	err := d.db.Get(&up,
		`SELECT
			user_id,
			username,
			realname,
			location,
			profile_uri,
			bio,
			social_1,
			social_2,
			social_3,
			created_at
		FROM 
			showcash.user
		WHERE username = $1`, handle,
	)

	up.Interests = []string{"MFA"}
	up.Friends = []UserProfile{}
	up.Followers = []UserProfile{}

	return up, err
}

func (d *DAO) getUserByUsernameAndPassword(username, password string) (User, error) {
	u := User{}
	err := d.db.Get(&u,
		`SELECT	
			user_id,
			username,
			realname,
			location,
			profile_uri,
			bio,
			social_1,
			social_2,
			social_3,
			email_address,
			password,
			shadow_banned
		FROM 
			showcash.user
		WHERE username = $1 AND password = $2`, username, password,
	)

	return u, err
}

func (d *DAO) createUser(u User) (User, error) {
	u.UserID = uuid.Must(uuid.NewV4())
	_, err := d.db.NamedExec(
		`INSERT INTO showcash.user(
			user_id,
			username,
			realname,
			location,
			profile_uri,
			bio,
			social_1,
			social_2,
			social_3,
			email_address,
			password
		) VALUES (
			:user_id,
			:username,
			:realname,
			:location,
			:profile_uri,
			:bio,
			:social_1,
			:social_2,
			:social_3,
			:email_address,
			:password
		)`, u,
	)
	return u, err
}

func (d *DAO) updateUser(u User) (User, error) {
	_, err := d.db.Exec(
		`UPDATE showcash.user SET
			realname   = $1,
			location   = $2,
			bio        = $3,
			social_1   = $4,
			social_2   = $5,
			social_3   = $6
			WHERE user_id = $7`,
		u.RealName, u.Location, u.Bio, u.Social1, u.Social2, u.Social3, u.UserID,
	)

	return u, err
}

// replaceSQL replaces the instance occurrence of any string pattern with an increasing $n based sequence
func replaceSQL(old, searchPattern string) string {
	tmpCount := strings.Count(old, searchPattern)
	for m := 1; m <= tmpCount; m++ {
		old = strings.Replace(old, searchPattern, "$"+strconv.Itoa(m), 1)
	}
	return old
}

// createTags is called every insert... it's a bit dirty
func (d *DAO) createTags(tags []string) ([]string, error) {

	// Sanitise tags on the way in as people
	// suck - return a nil so people can think they're amazing
	tags = cleanTags(tags)
	if len(tags) == 0 {
		return tags, nil
	}

	sqlStr := "INSERT INTO showcash.tag(tag) VALUES "
	queryParam := []interface{}{}
	for i := range tags {
		sqlStr += "(?),"
		queryParam = append(queryParam, tags[i])
	}

	//trim the last ,
	sqlStr = strings.TrimSuffix(sqlStr, ",")

	//Replacing ? with $n for postgres
	sqlStr = replaceSQL(sqlStr, "?")

	//prepare the statement
	stmt, err := d.db.Prepare(sqlStr + " ON CONFLICT DO NOTHING")

	//format all queryParam at once
	_, err = stmt.Exec(queryParam...)
	return tags, err
}

func (d *DAO) setPostTags(postID uuid.UUID, tags []string) {
	// Add batch of tags
	var err error
	if tags, err = d.createTags(tags); err != nil {
		log.Println("addTags().createTags() failed", err)
		return
	}
	// Now hook it up
	_, err = d.db.Exec(
		`INSERT INTO showcash.posttag(post_id,tag_id)
			SELECT $1, t.tag_id FROM showcash.tag AS t
			WHERE t.tag = ANY($2)
		ON CONFLICT DO NOTHING`,
		postID,
		pq.Array(tags),
	)
	if err != nil {
		log.Println("addTags() Failed", err)
	}
}

func (d *DAO) removePostTags(postID uuid.UUID, tags []string) {
	// remove a batch of tags
	_, err := d.db.Exec(
		`DELETE FROM showcash.posttag USING showcash.tag
			WHERE showcash.posttag.tag_id = showcash.tag.tag_id
				AND showcash.tag.tag = ANY($1) AND showcash.posttag.post_id = $2`,
		pq.Array(tags),
		// USING “ANY” INSTEAD OF “IN” with pq.Array() -- source:
		// https://www.opsdash.com/blog/postgres-arrays-golang.html
		postID,
	)
	if err != nil {
		log.Println("removeTags() Failed", err)
	}
}

func (d *DAO) getMostPopularTags() []string {
	var tags []string
	err := d.db.Select(
		&tags,
		`SELECT tag.tag FROM (
			SELECT t.tag,t.tag_id, COUNT(pt.*) AS pop FROM showcash.posttag AS pt
			JOIN showcash.tag AS t ON t.tag_id = pt.tag_id
			GROUP BY t.tag,t.tag_id ORDER BY pop DESC
		) AS p JOIN showcash.tag AS tag ON p.tag_id=tag.tag_id LIMIT 12`,
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println("getMostViewedPosts() failed", err)
	}
	return tags
}

func (d *DAO) getPostsByTags(tags []string) []Post {
	// NOTE: Only get by 1 tag at the moment
	var posts []Post
	err := d.db.Select(
		&posts,
		`SELECT DISTINCT (p.id),p.imageuri,p.title,p.date, u.username FROM showcash.post AS p
			JOIN showcash.user AS u ON p.user_id = u.user_id
			JOIN showcash.posttag AS pt ON pt.post_id = p.id
			JOIN showcash.tag AS t ON t.tag_id = pt.tag_id
		WHERE t.tag = ANY($1)
		ORDER BY p.id,p.date DESC
		LIMIT 50
		`, pq.Array(tags),
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println("getPostsByTags() failed", err)
	}
	return posts
}
