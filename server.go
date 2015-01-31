// Auth example is an example application which requires a login
// to view a private link. The username is "testuser" and the password
// is "password". This will require GORP and an SQLite3 database.
package main

import (
	"database/sql"
	"github.com/coopernurse/gorp"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessionauth"
	"github.com/martini-contrib/sessions"
	"log"
	"net/http"
	"os"
	"html/template"
	"bytes"
	//"fmt"
)

type TemplateData struct {
	Styles		[]string
	Users		*MyUserModel
	NavBarItems	[]NavBarItem
	Body		[]BodyItem
}

type NavBarItem struct {
	DefineName	string
	Info		NavBarData
}

type NavBarData struct {
	Label	string
	Class	string
	Link	string
	Glyph	string
}

type BodyItem struct {
	DefineName	string
	Info		BodyData
}

type BodyData struct {
	// Form Data?
}

var (
	dbmap	*gorp.DbMap
	templates *template.Template
)

func init() {
	templates, _ = template.ParseGlob("./templates/*/*.tmpl")
}

// initDb initialzes a new database structure
func initDb() *gorp.DbMap {
	db, err := sql.Open("mysql", "root@unix(/tmp/mysql.sock)/controlsurface")
	if err != nil {
		log.Fatalln("Fail to create database", err)
	}

	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}
	dbmap.AddTableWithName(MyUserModel{}, "users").SetKeys(true, "Id")

	return dbmap
}

func main() {
	// setup new session tracking
	store := sessions.NewCookieStore([]byte("ctrlsrfc"))
	dbmap = initDb()

	// initialize new martini Classic
	m := martini.Classic()
	m.Use(render.Renderer(render.Options{
		Directory: "templates",
		Layout: "layout",
		Funcs: []template.FuncMap {
			{
				"buildNavBarItems": func(args []NavBarItem) template.HTML {
					buf := bytes.NewBufferString("")
					for _, t := range args {
						err := templates.ExecuteTemplate(buf, t.DefineName, t.Info)
						if err != nil {
							return template.HTML("")
						}
					}
					return template.HTML(buf.String())
				},
				"buildBody": func(args []BodyItem) template.HTML {
					buf := bytes.NewBufferString("")
					for _, t := range args {
						err := templates.ExecuteTemplate(buf, t.DefineName, t.Info)
						if err != nil {
							return template.HTML("")
						}
					}
					return template.HTML(buf.String())
				},
			},
		},
	}))

	// Default our store to use Session cookies, so we don't leave logged in
	// users roaming around
	store.Options(sessions.Options{
		MaxAge: 0,
	})
	m.Use(sessions.Sessions("my_session", store))
	m.Use(sessionauth.SessionUser(GenerateAnonymousUser))
	sessionauth.RedirectUrl = "/login"
	sessionauth.RedirectParam = "next"

	m.Get("/", sessionauth.LoginRequired, func(r render.Render, user sessionauth.User) {
		var tmpldata TemplateData
		tmpldata.Users = user.(*MyUserModel)
		tmpldata.Styles = []string{"css/custom/sticky-footer-navbar.css"}
		tmpldata.NavBarItems = []NavBarItem{{DefineName: "std-nav-item", Info: NavBarData{Label: "Home", Class: "active", Glyph: "home", Link: "/"}}}
		tmpldata.NavBarItems = append(tmpldata.NavBarItems, NavBarItem{DefineName: "std-nav-item", Info: NavBarData{Label: "Domains", Class: "", Glyph: "cloud", Link: "/domains"}})
		tmpldata.NavBarItems = append(tmpldata.NavBarItems, NavBarItem{DefineName: "std-nav-item", Info: NavBarData{Label: "People", Class: "", Glyph: "user", Link: "/people"}})
		tmpldata.NavBarItems = append(tmpldata.NavBarItems, NavBarItem{DefineName: "std-nav-item", Info: NavBarData{Label: "Portfolios", Class: "", Glyph: "briefcase", Link: "/portfolios"}})
		tmpldata.Body = []BodyItem{{DefineName: "test-body"}}
		r.HTML(200, "index", tmpldata)
	})

	m.Get("/login", func(r render.Render) {
		tmpldata := TemplateData{Styles: []string{"css/custom/signin.css"}}
		r.HTML(200, "login", tmpldata)
	})

	m.Post("/login", binding.Bind(MyUserModel{}), func(session sessions.Session, postedUser MyUserModel, r render.Render, req *http.Request) {
		// You should verify credentials against a database or some other mechanism at this point.
		// Then you can authenticate this session.
		user := MyUserModel{}
		dbmap.TraceOn("[gorp]", log.New(os.Stdout, "myapp:", log.Lmicroseconds))
		err := dbmap.SelectOne(&user, "SELECT * FROM users WHERE username=? and password=?", postedUser.Username, postedUser.Password)
		if err != nil {
			r.Redirect(sessionauth.RedirectUrl)
			return
		} else {
			err := sessionauth.AuthenticateSession(session, &user)
			if err != nil {
				r.JSON(500, err)
			}

			params := req.URL.Query()
			redirect := params.Get(sessionauth.RedirectParam)
			r.Redirect(redirect)
			return
		}
	})

	m.Get("/domains", sessionauth.LoginRequired, func(r render.Render, user sessionauth.User) {
		var tmpldata TemplateData
		tmpldata.Users = user.(*MyUserModel)
		tmpldata.Styles = []string{"css/custom/sticky-footer-navbar.css"}
		tmpldata.NavBarItems = []NavBarItem{{DefineName: "std-nav-item", Info: NavBarData{Label: "Home", Class: "", Glyph: "home", Link: "/"}}}
		tmpldata.NavBarItems = append(tmpldata.NavBarItems, NavBarItem{DefineName: "std-nav-item", Info: NavBarData{Label: "Domains", Class: "active", Glyph: "cloud", Link: "/domains"}})
		tmpldata.NavBarItems = append(tmpldata.NavBarItems, NavBarItem{DefineName: "std-nav-item", Info: NavBarData{Label: "People", Class: "", Glyph: "user", Link: "/people"}})
		tmpldata.NavBarItems = append(tmpldata.NavBarItems, NavBarItem{DefineName: "std-nav-item", Info: NavBarData{Label: "Portfolios", Class: "", Glyph: "briefcase", Link: "/portfolios"}})
		r.HTML(200, "index", tmpldata)
	})

	m.Get("/people", sessionauth.LoginRequired, func(r render.Render, user sessionauth.User) {
		var tmpldata TemplateData
		tmpldata.Users = user.(*MyUserModel)
		tmpldata.Styles = []string{"css/custom/sticky-footer-navbar.css"}
		tmpldata.NavBarItems = []NavBarItem{{DefineName: "std-nav-item", Info: NavBarData{Label: "Home", Class: "", Glyph: "home", Link: "/"}}}
		tmpldata.NavBarItems = append(tmpldata.NavBarItems, NavBarItem{DefineName: "std-nav-item", Info: NavBarData{Label: "Domains", Class: "", Glyph: "cloud", Link: "/domains"}})
		tmpldata.NavBarItems = append(tmpldata.NavBarItems, NavBarItem{DefineName: "std-nav-item", Info: NavBarData{Label: "People", Class: "active", Glyph: "user", Link: "/people"}})
		tmpldata.NavBarItems = append(tmpldata.NavBarItems, NavBarItem{DefineName: "std-nav-item", Info: NavBarData{Label: "Portfolios", Class: "", Glyph: "briefcase", Link: "/portfolios"}})
		r.HTML(200, "index", tmpldata)
	})

	m.Get("/portfolios", sessionauth.LoginRequired, func(r render.Render, user sessionauth.User) {
		var tmpldata TemplateData
		tmpldata.Users = user.(*MyUserModel)
		tmpldata.Styles = []string{"css/custom/sticky-footer-navbar.css"}
		tmpldata.NavBarItems = []NavBarItem{{DefineName: "std-nav-item", Info: NavBarData{Label: "Home", Class: "", Glyph: "home", Link: "/"}}}
		tmpldata.NavBarItems = append(tmpldata.NavBarItems, NavBarItem{DefineName: "std-nav-item", Info: NavBarData{Label: "Domains", Class: "", Glyph: "cloud", Link: "/domains"}})
		tmpldata.NavBarItems = append(tmpldata.NavBarItems, NavBarItem{DefineName: "std-nav-item", Info: NavBarData{Label: "People", Class: "", Glyph: "user", Link: "/people"}})
		tmpldata.NavBarItems = append(tmpldata.NavBarItems, NavBarItem{DefineName: "std-nav-item", Info: NavBarData{Label: "Portfolios", Class: "active", Glyph: "briefcase", Link: "/portfolios"}})
		r.HTML(200, "index", tmpldata)
	})

	m.Get("/logout", sessionauth.LoginRequired, func(session sessions.Session, user sessionauth.User, r render.Render) {
		sessionauth.Logout(session, user)
		r.Redirect("/")
	})

	m.Run()
}