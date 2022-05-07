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
        "/v1/auth/createUser": {
            "post": {
                "description": "AuthingCreateUser",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Authing"
                ],
                "summary": "AuthingCreateUser",
                "parameters": [
                    {
                        "description": "body for user info",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.CreateUserInput"
                        }
                    }
                ],
                "responses": {}
            }
        },
        "/v1/auth/getDetail/{authingUserId}": {
            "get": {
                "description": "AuthingGetToken",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Authing"
                ],
                "summary": "AuthingGetToken",
                "parameters": [
                    {
                        "type": "string",
                        "description": "The key for staticblock",
                        "name": "authingUserId",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {}
            }
        },
        "/v1/auth/loginok": {
            "get": {
                "description": "login success redirect url",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Authing"
                ],
                "summary": "login success redirect url",
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
                    "v1 version"
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
                    "v1 version"
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
        "/v1/images/queryHistory/mine": {
            "get": {
                "description": "Query My History",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "v1 version"
                ],
                "summary": "QueryMyHistory",
                "parameters": [
                    {
                        "type": "string",
                        "description": "arch",
                        "name": "arch",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "status",
                        "name": "status",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "build type",
                        "name": "type",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "name or desc",
                        "name": "nameordesc",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "offset ",
                        "name": "offset",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "limit",
                        "name": "limit",
                        "in": "query"
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
                    "v1 version"
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
                    "v1 version"
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
                    "v1 version"
                ],
                "summary": "StartBuild Job",
                "parameters": [
                    {
                        "description": "body for ImageMeta content",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.BuildParam"
                        }
                    }
                ],
                "responses": {}
            }
        },
        "/v2/images/createJob": {
            "post": {
                "description": "start a image build job",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "v2 version"
                ],
                "summary": "Create Job",
                "parameters": [
                    {
                        "description": "body for ImageMeta content",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.BuildParam"
                        }
                    }
                ],
                "responses": {}
            }
        },
        "/v2/images/deleteJob": {
            "post": {
                "description": "delete multipule job build records",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "v2 version"
                ],
                "summary": "deleteRecord",
                "parameters": [
                    {
                        "description": "job id list",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "string"
                            }
                        }
                    }
                ],
                "responses": {}
            }
        },
        "/v2/images/getJobParam/{id}": {
            "get": {
                "description": "get job build param",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "v2 version"
                ],
                "summary": "GetJobParam",
                "parameters": [
                    {
                        "type": "string",
                        "description": "job id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {}
            }
        },
        "/v2/images/getLogsOf/{id}": {
            "get": {
                "description": "get single job logs",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "v2 version"
                ],
                "summary": "get single job logs",
                "parameters": [
                    {
                        "type": "string",
                        "description": "job id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "step id",
                        "name": "stepID",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "uuid",
                        "name": "uuid",
                        "in": "query"
                    }
                ],
                "responses": {}
            }
        },
        "/v2/images/getMySummary": {
            "get": {
                "description": "get my summary",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "v2 version"
                ],
                "summary": "MySummary",
                "responses": {}
            }
        },
        "/v2/images/getOne/{id}": {
            "get": {
                "description": "get single job detail",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "v2 version"
                ],
                "summary": "get single job detail",
                "parameters": [
                    {
                        "type": "string",
                        "description": "job id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {}
            }
        },
        "/v2/images/stopJob/{id}": {
            "delete": {
                "description": "Stop Job Build",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "v2 version"
                ],
                "summary": "StopJobBuild",
                "parameters": [
                    {
                        "type": "string",
                        "description": "job id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {}
            }
        },
        "/v3/baseImages/import": {
            "post": {
                "description": "import  a image meta data",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "v3 version"
                ],
                "summary": "ImportBaseImages",
                "parameters": [
                    {
                        "description": "body for BaseImages content",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.BaseImages"
                        }
                    }
                ],
                "responses": {}
            }
        },
        "/v3/baseImages/list": {
            "get": {
                "description": "get my base image list order by id desc",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "v3 version"
                ],
                "summary": "ListBaseImages",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "offset ",
                        "name": "offset",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "limit",
                        "name": "limit",
                        "in": "query"
                    }
                ],
                "responses": {}
            }
        },
        "/v3/baseImages/{id}": {
            "put": {
                "description": "update  a base  images data",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "v3 version"
                ],
                "summary": "UpdateBaseImages",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "id for  content",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "body for BaseImages content",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.BaseImages"
                        }
                    }
                ],
                "responses": {}
            },
            "delete": {
                "description": "delete  a base  images data",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "v3 version"
                ],
                "summary": "DeletBaseImages",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "id for BaseImages content",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {}
            }
        },
        "/v3/getImagesAndKickStart": {
            "get": {
                "description": "GetImagesAndKickStart",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "v3 version"
                ],
                "summary": "GetImagesAndKickStart",
                "responses": {}
            }
        },
        "/v3/images/buildFromIso": {
            "post": {
                "description": "build a image from iso",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "v3 version"
                ],
                "summary": "BuildFromISO",
                "parameters": [
                    {
                        "description": "body for ImageMeta content",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.BaseImagesKickStart"
                        }
                    }
                ],
                "responses": {}
            }
        },
        "/v3/kickStart": {
            "post": {
                "description": "add  a KickStart data",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "v3 version"
                ],
                "summary": "AddKickStart",
                "parameters": [
                    {
                        "type": "file",
                        "description": "kickstart file",
                        "name": "file",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "  name",
                        "name": "name",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "  desc",
                        "name": "desc",
                        "in": "formData",
                        "required": true
                    }
                ],
                "responses": {}
            }
        },
        "/v3/kickStart/list": {
            "get": {
                "description": "get my kick start file list order by id desc",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "v3 version"
                ],
                "summary": "ListKickStart",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "offset ",
                        "name": "offset",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "limit",
                        "name": "limit",
                        "in": "query"
                    }
                ],
                "responses": {}
            }
        },
        "/v3/kickStart/{id}": {
            "get": {
                "description": "GetKickStartByID",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "v3 version"
                ],
                "summary": "GetKickStartByID",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "id for  content",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {}
            },
            "put": {
                "description": "update  a kick start data",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "v3 version"
                ],
                "summary": "UpdateKickStart",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "id for  content",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "body for KickStart content",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.KickStart"
                        }
                    }
                ],
                "responses": {}
            },
            "delete": {
                "description": "delete  a KickStart data",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "v3 version"
                ],
                "summary": "DeleteKickStart",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "id for KickStart content",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {}
            }
        }
    },
    "definitions": {
        "models.BaseImages": {
            "type": "object",
            "properties": {
                "arch": {
                    "type": "string"
                },
                "checksum": {
                    "type": "string"
                },
                "createTime": {
                    "type": "string"
                },
                "desc": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "name": {
                    "type": "string"
                },
                "status": {
                    "type": "string"
                },
                "url": {
                    "type": "string"
                },
                "userId": {
                    "type": "integer"
                }
            }
        },
        "models.BaseImagesKickStart": {
            "type": "object",
            "properties": {
                "baseImageID": {
                    "type": "string"
                },
                "desc": {
                    "type": "string"
                },
                "kickStartContent": {
                    "type": "string"
                },
                "kickStartName": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                }
            }
        },
        "models.BuildParam": {
            "type": "object",
            "properties": {
                "arch": {
                    "description": "Id        int      ` + "`" + `gorm:\"primaryKey\"` + "`" + `",
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
                "desc": {
                    "type": "string"
                },
                "label": {
                    "type": "string"
                },
                "release": {
                    "type": "string"
                }
            }
        },
        "models.CreateUserInput": {
            "type": "object",
            "properties": {
                "address": {
                    "type": "string"
                },
                "birthdate": {
                    "type": "string"
                },
                "blocked": {
                    "type": "boolean"
                },
                "browser": {
                    "type": "string"
                },
                "company": {
                    "type": "string"
                },
                "country": {
                    "type": "string"
                },
                "device": {
                    "type": "string"
                },
                "email": {
                    "type": "string"
                },
                "emailVerified": {
                    "type": "boolean"
                },
                "externalId": {
                    "type": "string"
                },
                "familyName": {
                    "type": "string"
                },
                "formatted": {
                    "type": "string"
                },
                "gender": {
                    "type": "string"
                },
                "givenName": {
                    "type": "string"
                },
                "isDeleted": {
                    "type": "boolean"
                },
                "lastIP": {
                    "type": "string"
                },
                "lastLogin": {
                    "type": "string"
                },
                "locale": {
                    "type": "string"
                },
                "locality": {
                    "type": "string"
                },
                "loginsCount": {
                    "type": "integer"
                },
                "middleName": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "nickname": {
                    "type": "string"
                },
                "oauth": {
                    "type": "string"
                },
                "openid": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                },
                "phone": {
                    "type": "string"
                },
                "phoneVerified": {
                    "type": "boolean"
                },
                "photo": {
                    "type": "string"
                },
                "postalCode": {
                    "type": "string"
                },
                "preferredUsername": {
                    "type": "string"
                },
                "profile": {
                    "type": "string"
                },
                "region": {
                    "type": "string"
                },
                "registerSource": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "signedUp": {
                    "type": "string"
                },
                "streetAddress": {
                    "type": "string"
                },
                "unionid": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                },
                "website": {
                    "type": "string"
                },
                "zoneinfo": {
                    "type": "string"
                }
            }
        },
        "models.KickStart": {
            "type": "object",
            "properties": {
                "content": {
                    "type": "string"
                },
                "createTime": {
                    "type": "string"
                },
                "desc": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "name": {
                    "type": "string"
                },
                "updateTime": {
                    "type": "string"
                },
                "userId": {
                    "type": "integer"
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
