package main

import (
	"log"

	"project_crud_with_auth_tmpl/components"
	"project_crud_with_auth_tmpl/layouts"
	"project_crud_with_auth_tmpl/models"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/", helloWorldHandler)
	e.GET("/projects", listProjectsHandler)

	// Start server
	log.Fatal(e.Start(":8080"))
}

func helloWorldHandler(c echo.Context) error {

	// リクエストがhtmxからの場合、コンポーネントのみをレンダリングする。

	if c.Request().Header.Get("HX-Request") == "true" {

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)

		return components.HelloWorld().Render(c.Request().Context(), c.Response().Writer)

	}

	// それ以外の場合、完全なレイアウトをレンダリングする

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)

	// Hello World用には、一貫性のためにBaseレイアウトでラップする。

	return layouts.Base("Hello", components.HelloWorld()).Render(c.Request().Context(), c.Response().Writer)

}

func listProjectsHandler(c echo.Context) error {

	// モックデータ

	projects := []models.Project{

		{ID: 1, Name: "プロジェクト Alpha", Description: "最初のプロジェクトです", Status: "進行中"},

		{ID: 2, Name: "プロジェクト Beta", Description: "2番目のプロジェクトです", Status: "保留中"},

		{ID: 3, Name: "プロジェクト Gamma", Description: "3番目のプロジェクトです", Status: "完了"},
	}

	// 1. コンポーネントの準備

	content := components.ProjectList(projects)

	// 2. デュアルモードレンダリング

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)

	if c.Request().Header.Get("HX-Request") == "true" {

		// リストコンポーネント（フラグメント）のみをレンダリング

		return content.Render(c.Request().Context(), c.Response().Writer)

	}

	// コンポーネントをラップした完全なレイアウトをレンダリング

	return layouts.Base("プロジェクト一覧", content).Render(c.Request().Context(), c.Response().Writer)

}
