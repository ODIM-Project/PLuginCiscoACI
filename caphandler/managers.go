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
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/ODIM-Project/PluginCiscoACI/capmodel"
	"github.com/ODIM-Project/PluginCiscoACI/capresponse"
	"github.com/ODIM-Project/PluginCiscoACI/caputilities"
	pluginConfig "github.com/ODIM-Project/PluginCiscoACI/config"
	iris "github.com/kataras/iris/v12"
)

//GetManagersCollection  Fetches details of the given resource from the device
func GetManagersCollection(ctx iris.Context) {
	//Get token from Request
	token := ctx.GetHeader("X-Auth-Token")
	uri := ctx.Request().RequestURI
	//Validating the token
	if token != "" {
		flag := TokenValidation(token)
		if !flag {
			log.Println("Invalid/Expired X-Auth-Token")
			ctx.StatusCode(http.StatusUnauthorized)
			ctx.WriteString("Invalid/Expired X-Auth-Token")
			return
		}
	}
	var deviceDetails capmodel.Device
	// if any error come while getting the device then request will be for  plugins manager
	ctx.ReadJSON(&deviceDetails)
	if deviceDetails.Host == "" {
		var members = []capresponse.Link{
			capresponse.Link{
				Oid: "/ODIM/v1/Managers/" + pluginConfig.Data.RootServiceUUID,
			},
		}

		managers := capresponse.ManagersCollection{
			OdataContext: "/ODIM/v1/$metadata#ManagerCollection.ManagerCollection",
			//Etag:         "W/\"AA6D42B0\"",
			OdataID:      uri,
			OdataType:    "#ManagerCollection.ManagerCollection",
			Description:  "Managers view",
			Name:         "Managers",
			Members:      members,
			MembersCount: len(members),
		}
		ctx.StatusCode(http.StatusOK)
		ctx.JSON(managers)
		return
	}
	getInfoFromDevice(uri, deviceDetails, ctx)
	return

}

//GetManagersInfo Fetches details of the given resource from the device
func GetManagersInfo(ctx iris.Context) {
	//Get token from Request
	token := ctx.GetHeader("X-Auth-Token")
	uri := ctx.Request().RequestURI

	//Validating the token
	if token != "" {
		flag := TokenValidation(token)
		if !flag {
			log.Println("Invalid/Expired X-Auth-Token")
			ctx.StatusCode(http.StatusUnauthorized)
			ctx.WriteString("Invalid/Expired X-Auth-Token")
			return
		}
	}
	var deviceDetails capmodel.Device
	// if any error come while getting the device then request will be for  plugins manager
	ctx.ReadJSON(&deviceDetails)
	if deviceDetails.Host == "" {
		managers := capresponse.Manager{
			OdataContext: "/ODIM/v1/$metadata#Manager.Manager",
			//Etag:            "W/\"AA6D42B0\"",
			OdataID:         uri,
			OdataType:       "#Manager.v1_3_3.Manager",
			Name:            pluginConfig.Data.PluginConf.ID,
			ManagerType:     "Service",
			ID:              pluginConfig.Data.RootServiceUUID,
			UUID:            pluginConfig.Data.RootServiceUUID,
			FirmwareVersion: pluginConfig.Data.FirmwareVersion,
			Status: &capresponse.ManagerStatus{
				State: "Enabled",
			},
		}
		ctx.StatusCode(http.StatusOK)
		ctx.JSON(managers)
		return
	}
	getInfoFromDevice(uri, deviceDetails, ctx)
	return

}

func getInfoFromDevice(uri string, deviceDetails capmodel.Device, ctx iris.Context) {
	//replacing the request url with south bound translation URL
	for key, value := range pluginConfig.Data.URLTranslation.SouthBoundURL {
		uri = strings.Replace(uri, key, value, -1)
	}
	device := &caputilities.RedfishDevice{
		Host:     deviceDetails.Host,
		Username: deviceDetails.Username,
		Password: string(deviceDetails.Password),
	}
	redfishClient, err := caputilities.GetRedfishClient()
	if err != nil {
		errMsg := "error: internal processing error: " + err.Error()
		log.Println(errMsg)
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.WriteString(errMsg)
		return
	}

	//Fetching generic resource details from the device
	resp, err := redfishClient.GetWithBasicAuth(device, uri)
	if err != nil {
		errMsg := "error: authentication failed: " + err.Error()
		log.Println(errMsg)
		if resp == nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.WriteString(errMsg)
			return
		}
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf(err.Error())
	}

	if resp.StatusCode == 401 {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.WriteString("Authtication with the device failed")
		return
	}
	if resp.StatusCode >= 300 {
		fmt.Printf("Could not retreive generic resource for %s: \n%s\n\n", device.Host, body)
	}
	respData := string(body)
	//replacing the resposne with north bound translation URL
	for key, value := range pluginConfig.Data.URLTranslation.NorthBoundURL {
		respData = strings.Replace(respData, key, value, -1)
	}
	ctx.StatusCode(resp.StatusCode)
	ctx.Write([]byte(respData))
}
