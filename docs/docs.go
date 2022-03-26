// Package docs GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/v1/images/delete/:id": {
            "delete": {
                "description": "update single data",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "meta Manager"
                ],
                "summary": "delete",
                "parameters": [
                    {
                        "type": "string",
                        "description": "The key for staticblock",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {}
            }
        },
        "/v1/images/get/{id}": {
            "get": {
                "description": "get single one",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "meta Manager"
                ],
                "summary": "get",
                "parameters": [
                    {
                        "type": "string",
                        "description": "The key for staticblock",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {}
            }
        },
        "/v1/images/param/getBaseData/": {
            "get": {
                "description": "get architecture, release Version, output Format ,and default package name list",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "meta Manager"
                ],
                "summary": "GetBaseData param",
                "responses": {}
            }
        },
        "/v1/images/param/getCustomePkgList/": {
            "get": {
                "description": "get custom package name list",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "meta Manager"
                ],
                "summary": "GetCustomePkgList param",
                "parameters": [
                    {
                        "type": "string",
                        "description": " arch ,e g:x86_64",
                        "name": "arch",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "release  ",
                        "name": "release",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "custom group  ",
                        "name": "sig",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {}
            }
        },
        "/v1/images/query": {
            "get": {
                "description": "use param to query multi datas",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "meta Manager"
                ],
                "summary": "query multi datas",
                "parameters": [
                    {
                        "type": "string",
                        "description": "project name",
                        "name": "project_name",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "package name",
                        "name": "pkg_name",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {}
            }
        },
        "/v1/images/queryJobLogs/{name}": {
            "get": {
                "description": "QueryJobLogs for given job name",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "meta Manager"
                ],
                "summary": "QueryJobLogs",
                "parameters": [
                    {
                        "type": "string",
                        "description": "The name for job",
                        "name": "name",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {}
            }
        },
        "/v1/images/queryJobStatus/{name}": {
            "get": {
                "description": "QueryJobStatus for given job name",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "meta Manager"
                ],
                "summary": "QueryJobStatus",
                "parameters": [
                    {
                        "type": "string",
                        "description": "The name for job",
                        "name": "name",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "The id for job in database. ",
                        "name": "id",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "job namespace ",
                        "name": "ns",
                        "in": "query"
                    }
                ],
                "responses": {}
            }
        },
        "/v1/images/startBuild": {
            "post": {
                "description": "start a image build job",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "meta Manager"
                ],
                "summary": "StartBuild Job",
                "parameters": [
                    {
                        "description": "body for ImageMeta content",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.ImageInputData"
                        }
                    }
                ],
                "responses": {}
            }
        }
    },
    "definitions": {
        "models.ImageInputData": {
            "type": "object",
            "properties": {
                "arch": {
                    "type": "string"
                },
                "buildType": {
                    "type": "string"
                },
                "customPkg": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "id": {
                    "type": "integer"
                },
                "release": {
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
