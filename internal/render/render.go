package render

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/justinas/nosurf"
	"github.com/rest5387/myApp1/internal/config"
	"github.com/rest5387/myApp1/internal/models"
)

var functions = template.FuncMap{}
var app *config.AppConfig
var pathToTemplates = "./templates"

//NewTemplates sets the config for the template package
func NewTemplates(a *config.AppConfig) {
	app = a
}

// AddDefaultData adds data for all templates
func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Error = app.Session.PopString(r.Context(), "error")
	td.Warning = app.Session.PopString(r.Context(), "warning")
	td.CSRFToken = nosurf.Token(r)
	return td
}

//RenderTemplate renders templates using html/template
func RenderTemplate(w http.ResponseWriter, r *http.Request, tmpl string, td *models.TemplateData) error {

	var tc map[string]*template.Template
	if app.UseCache {
		//get the template cache from app config

		tc = app.TemplateCache
	} else {
		tc, _ = CreateTemplateCache()
	}
	// tc, err := CreateTemplateCache()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	t, ok := tc[tmpl]
	// fmt.Println("tc[", tmpl, "]", tc[tmpl].DefinedTemplates())
	if !ok {
		return errors.New("can't get template from template cache")
	}

	buf := new(bytes.Buffer)

	td = AddDefaultData(td, r)

	_ = t.Execute(buf, td)

	_, err := buf.WriteTo(w)
	if err != nil {
		fmt.Println("Error writing template to browser", err)
		return err
	}

	return nil
	// parsedTemplate, _ := template.ParseFiles("./templates/" + tmpl)
	// err = parsedTemplate.Execute(w, nil)
	// if err != nil {
	// 	fmt.Println("Error parsing template: ", err)
	// 	return
	// }
}

//CreateTemplateCache creates a template cache as a map
func CreateTemplateCache() (map[string]*template.Template, error) {
	myCache := map[string]*template.Template{}

	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl", pathToTemplates))

	if err != nil {
		return myCache, err
	}

	for _, page := range pages {
		// fmt.Println("Page is currently", page)
		name := filepath.Base(page)
		// fmt.Println("Page is currently", name)
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)

		if err != nil {
			return myCache, err
		}

		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
		if err != nil {
			return myCache, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
			if err != nil {
				return myCache, err
			}
		}
		myCache[name] = ts
	}
	return myCache, nil
}