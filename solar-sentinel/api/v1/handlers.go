package v1

import (
	"net/http"
	"github.com/negeek/solar-sphere/solar-sentinel/utils"
	model"github.com/negeek/solar-sphere/solar-sentinel/repository/v1"
)

func Create_DeviceID(w http.ResponseWriter, r *http.Request){

	var (
		device = &model.Device{}
		err error
	)
	// Read  request body
	err= utils.Unmarshall(r.Body, device)
	if err != nil{
		utils.JsonResponse(w, false, http.StatusBadRequest , err.Error(), nil)
		return	
	}

	device.ID=utils.GenerateID()

	// Create device
	err=device.Create()
	if err != nil{
		utils.JsonResponse(w, false, http.StatusBadRequest , "Error creating device", nil)
		return	
	}

	// Response
	utils.JsonResponse(w, true, http.StatusCreated ,"Successfully created device", map[string]interface{}{"device_id":device.ID})
	return	
}