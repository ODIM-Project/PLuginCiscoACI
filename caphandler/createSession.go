//(C) Copyright [2020] Hewlett Packard Enterprise Development LP
//
//Licensed under the Apache License, Version 2.0 (the "License"); you may
//not use this file except in compliance with the License. You may obtain
//a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
//WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
//License for the specific language governing permissions and limitations
// under the License.

//Package caphandler ...
package caphandler

import (
	"encoding/base64"
	"encoding/json"
	"github.com/ODIM-Project/PluginCiscoACI/capmodel"
	pluginConfig "github.com/ODIM-Project/PluginCiscoACI/config"
	iris "github.com/kataras/iris/v12"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/sha3"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

//TokenMap is used to define the plugin generated tokens
type TokenMap struct {
	Token    string
	LastUsed time.Time
}

var tokenDetails []TokenMap
var tokenSpec TokenMap

//CreateSession is used to create session for odimra to interact with plugin
func CreateSession(ctx iris.Context) {
	var userCreds capmodel.Users
	rawBodyAsBytes, err := ioutil.ReadAll(ctx.Request().Body)
	if err != nil {
		errorMessage := "Error while trying to validate the credentials: " + err.Error()
		log.Println(errorMessage)
		ctx.StatusCode(http.StatusBadRequest)
		ctx.WriteString(errorMessage)
	}
	err = json.Unmarshal(rawBodyAsBytes, &userCreds)
	if err != nil {
		errorMessage := "Error while trying to unmarshal user details: " + err.Error()
		log.Println(errorMessage)
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.WriteString(errorMessage)
	}
	//Validate the credentials
	userName := userCreds.Username
	password := userCreds.Password
	validateResponse := validate(userName, password)
	if !validateResponse {
		errorMessage := "Invalid credentials for session creation"
		log.Println(errorMessage)
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.WriteString(errorMessage)
		return
	}
	//Create token
	token := createToken()
	currentTime := time.Now()
	tokenSpec.Token = token
	tokenSpec.LastUsed = currentTime
	tokenDetails = append(tokenDetails, tokenSpec)
	ctx.StatusCode(http.StatusCreated)
	ctx.Header("X-Auth-Token", token)
}

func validate(userName, password string) bool {
	//var err error
	username := pluginConfig.Data.PluginConf.UserName
	passwd := pluginConfig.Data.PluginConf.Password
	if username != userName {
		return false
	}
	hash := sha3.New512()
	hash.Write([]byte(password))
	hashSum := hash.Sum(nil)
	hashedPassword := base64.URLEncoding.EncodeToString(hashSum)
	if passwd != hashedPassword {
		log.Println("username/password does not match")
		return false
	}

	return true
}

func createToken() string {
	token := uuid.NewV4().String()
	return token
}
