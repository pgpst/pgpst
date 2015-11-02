package api

import (
	"strconv"
	"strings"
	"time"

	r "github.com/pgpst/pgpst/internal/github.com/dancannon/gorethink"
	"github.com/pgpst/pgpst/internal/github.com/dchest/uniuri"
	"github.com/pgpst/pgpst/internal/github.com/gin-gonic/gin"

	"github.com/pgpst/pgpst/pkg/models"
)

func (a *API) createResource(c *gin.Context) {
	// Get token and account info from the context
	var (
		account = c.MustGet("account").(*models.Account)
		token   = c.MustGet("token").(*models.Token)
	)

	// Check the scope
	if !models.InScope(token.Scope, []string{"resources:create"}) {
		c.JSON(403, &gin.H{
			"code":  0,
			"error": "Your token has insufficient scope",
		})
		return
	}

	// Decode the input
	var input struct {
		Meta map[string]interface{} `json:"meta"`
		Body []byte                 `json:"body"`
		Tags []string               `json:"tags"`
	}
	if err := c.Bind(&input); err != nil {
		c.JSON(422, &gin.H{
			"code":    0,
			"message": err.Error(),
		})
		return
	}

	// Save it into database
	resource := &models.Resource{
		ID:           uniuri.NewLen(uniuri.UUIDLen),
		DateCreated:  time.Now(),
		DateModified: time.Now(),
		Owner:        account.ID,

		Meta: input.Meta,
		Body: input.Body,
		Tags: input.Tags,
	}

	// Insert it into database
	if err := r.Table("resources").Insert(resource).Exec(a.Rethink); err != nil {
		c.JSON(500, &gin.H{
			"code":    0,
			"message": err.Error(),
		})
		return
	}

	c.JSON(201, resource)
}

func (a *API) getAccountResources(c *gin.Context) {
	// Token and account from context
	var (
		ownAccount = c.MustGet("account").(*models.Account)
		token      = c.MustGet("token").(*models.Token)
	)

	// Resolve the ID from the URL
	id := c.Param("id")
	if id == "me" {
		id = ownAccount.ID
	}

	// Check the scope
	if id == ownAccount.ID {
		if !models.InScope(token.Scope, []string{"resources:read"}) {
			c.JSON(403, &gin.H{
				"code":  0,
				"error": "Your token has insufficient scope",
			})
			return
		}
	} else {
		if !models.InScope(token.Scope, []string{"admin"}) {
			c.JSON(403, &gin.H{
				"code":  0,
				"error": "Your token has insufficient scope",
			})
			return
		}
	}

	// 1. Owner filter
	query := r.Table("resources").GetAllByIndex("owner", id).Without("body")

	// 2. Tag filter
	if tagstr := c.Query("tags"); tagstr != "" {
		tags := strings.Split(tagstr, ",")

		// Cast to []interface{}
		tagsi := []interface{}{}
		for _, tag := range tags {
			tagsi = append(tagsi, tag)
		}

		query = query.Filter(func(row r.Term) r.Term {
			return row.Field("tags").Contains(tagsi...)
		})
	}

	// 3. Meta filter
	// not needed right now

	// 4. Date created and date modified
	ts := func(field string) error {
		if dm := c.Query(field); dm != "" {
			dmp := strings.Split(dm, ",")
			if len(dmp) == 1 || dmp[1] == "" {
				// parse dmp[0]
				d0, err := time.Parse(time.RFC3339, dmp[0])
				if err != nil {
					return err
				}

				// after dmp[0]
				query = query.Filter(func(row r.Term) r.Term {
					return row.Field(field).Ge(d0)
				})
			} else {
				// parse dmp[1]
				d1, err := time.Parse(time.RFC3339, dmp[1])
				if err != nil {
					return err
				}

				if dmp[0] == "" {
					// until dmp[1]
					query = query.Filter(func(row r.Term) r.Term {
						return row.Field(field).Le(d1)
					})
				} else {
					// parse dmp[0]
					d0, err := time.Parse(time.RFC3339, dmp[0])
					if err != nil {
						return err
					}

					// between dmp[0] and dmp[1]
					query = query.Filter(func(row r.Term) r.Term {
						return row.Field(field).Ge(d0).And(row.Field(field).Le(d1))
					})
				}
			}
		}

		return nil
	}
	if err := ts("date_modified"); err != nil {
		c.JSON(500, &gin.H{
			"code":  0,
			"error": err.Error(),
		})
		return
	}
	if err := ts("date_created"); err != nil {
		c.JSON(500, &gin.H{
			"code":  0,
			"error": err.Error(),
		})
		return
	}

	// 5. Pluck / Without
	if pls := c.Query("pluck"); pls != "" {
		pl := strings.Split(pls, ",")
		// Cast to []interface{}
		pli := []interface{}{}
		for _, field := range pl {
			pli = append(pli, field)
		}
		query = query.Pluck(pli...)
	} else if wos := c.Query("wo"); wos != "" {
		wo := strings.Split(wos, ",")
		// Cast to []interface{}
		woi := []interface{}{}
		for _, field := range wo {
			woi = append(woi, field)
		}
		query = query.Pluck(woi...)
	}

	// 6. Ordering
	if obs := c.Query("order_by"); obs != "" {
		ob := strings.Split(obs, ",")
		fields := []interface{}{}
		for _, fi := range ob {
			asc := true
			if fi[0] == '-' {
				asc = false
				fi = fi[1:]
			} else if fi[0] == '+' || fi[0] == ' ' {
				fi = fi[1:]
			}

			field := r.Row.Field(fi)

			if path := strings.Split(fi, "."); len(path) > 1 {
				field = r.Row.Field(path[0])

				for i := 1; i < len(path); i++ {
					field = field.Field(path[i])
				}
			}

			if !asc {
				field = r.Desc(field)
			}

			fields = append(fields, field)
		}

		query = query.OrderBy(fields...)
	}

	// 7. Limiting
	var (
		sks         = c.Query("skip")
		lms         = c.Query("limit")
		err         error
		skip, limit int
	)
	if sks != "" {
		skip, err = strconv.Atoi(sks)
		if err != nil {
			c.JSON(400, &gin.H{
				"code":  0,
				"error": err.Error(),
			})
			return
		}
	}

	if lms != "" {
		limit, err = strconv.Atoi(lms)
		if err != nil {
			c.JSON(400, &gin.H{
				"code":  0,
				"error": err.Error(),
			})
			return
		}
	}

	if skip != 0 && limit != 0 {
		query = query.Slice(skip, skip+limit)
	} else if skip == 0 && limit != 0 {
		query = query.Limit(limit)
	} else if skip != 0 && limit == 0 {
		query = query.Skip(skip)
	}

	// Get resources from database without bodies
	cursor, err := query.Run(a.Rethink)
	if err != nil {
		c.JSON(500, &gin.H{
			"code":  0,
			"error": err.Error(),
		})
		return
	}
	defer cursor.Close()
	var resources []*models.Resource
	if err := cursor.All(&resources); err != nil {
		c.JSON(500, &gin.H{
			"code":  0,
			"error": err.Error(),
		})
		return
	}

	// Write the response
	c.JSON(200, resources)
	return
}

func (a *API) readResource(c *gin.Context) {
	// Get token and account info from the context
	var (
		account = c.MustGet("account").(*models.Account)
		token   = c.MustGet("token").(*models.Token)
	)

	// Resolve the resource ID and fetch it from database
	id := c.Param("id")
	cursor, err := r.Table("resources").Get(id).Run(a.Rethink)
	if err != nil {
		c.JSON(500, &gin.H{
			"code":  0,
			"error": err.Error(),
		})
		return
	}
	defer cursor.Close()
	var resource *models.Resource
	if err := cursor.All(&resource); err != nil {
		c.JSON(500, &gin.H{
			"code":  0,
			"error": err.Error(),
		})
		return
	}

	if resource.Owner == account.ID {
		// Check the scope
		if !models.InScope(token.Scope, []string{"resources:read"}) {
			c.JSON(403, &gin.H{
				"code":  0,
				"error": "Your token has insufficient scope",
			})
			return
		}
	} else {
		// Check the scope
		if !models.InScope(token.Scope, []string{"admin"}) {
			c.JSON(403, &gin.H{
				"code":  0,
				"error": "Your token has insufficient scope",
			})
			return
		}
	}

	c.JSON(200, resource)
}
