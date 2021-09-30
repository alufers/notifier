// Package docs GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag
package docs

import (
	"bytes"
	"encoding/json"
	"strings"
	"text/template"

	"github.com/swaggo/swag"
)

var doc = `{
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
        "/notify": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Delivers a notification to all the sinks",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Send a notification",
                "operationId": "post-notification",
                "parameters": [
                    {
                        "description": "Notification to deliver",
                        "name": "notification",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/notifier.PostNotifyBody"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/notifier.PostNotifyResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/notifier.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/question": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Currently supported question types: yesno",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Asks a question to the user",
                "operationId": "post-question",
                "parameters": [
                    {
                        "description": "Question to ask",
                        "name": "notification",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/notifier.PostQuestionBody"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/notifier.PostQuestionResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/notifier.ErrorResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "notifier.Answer": {
            "type": "object",
            "properties": {
                "answerDuration": {
                    "type": "integer"
                },
                "timedOut": {
                    "type": "boolean"
                },
                "value": {}
            }
        },
        "notifier.ErrorResponse": {
            "type": "object",
            "properties": {
                "error": {
                    "type": "string"
                }
            }
        },
        "notifier.PostNotifyBody": {
            "type": "object",
            "properties": {
                "body": {
                    "type": "string"
                },
                "title": {
                    "type": "string"
                }
            }
        },
        "notifier.PostNotifyResponse": {
            "type": "object",
            "properties": {
                "deliveriesCucceeded": {
                    "type": "integer"
                },
                "deliveriesTotal": {
                    "type": "integer"
                },
                "errors": {
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                }
            }
        },
        "notifier.PostQuestionBody": {
            "type": "object",
            "properties": {
                "kind": {
                    "type": "string"
                },
                "text": {
                    "type": "string"
                },
                "timeout": {
                    "type": "string"
                }
            }
        },
        "notifier.PostQuestionResponse": {
            "type": "object",
            "properties": {
                "answer": {
                    "$ref": "#/definitions/notifier.Answer"
                },
                "errors": {
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                }
            }
        }
    }
}`

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = swaggerInfo{
	Version:     "",
	Host:        "",
	BasePath:    "",
	Schemes:     []string{},
	Title:       "",
	Description: "",
}

type s struct{}

func (s *s) ReadDoc() string {
	sInfo := SwaggerInfo
	sInfo.Description = strings.Replace(sInfo.Description, "\n", "\\n", -1)

	t, err := template.New("swagger_info").Funcs(template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
		"escape": func(v interface{}) string {
			// escape tabs
			str := strings.Replace(v.(string), "\t", "\\t", -1)
			// replace " with \", and if that results in \\", replace that with \\\"
			str = strings.Replace(str, "\"", "\\\"", -1)
			return strings.Replace(str, "\\\\\"", "\\\\\\\"", -1)
		},
	}).Parse(doc)
	if err != nil {
		return doc
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, sInfo); err != nil {
		return doc
	}

	return tpl.String()
}

func init() {
	swag.Register(swag.Name, &s{})
}
