package v1

import (
	"os"
	"fmt"
	"net/http"
	"github.com/negeek/solar-sphere/solar-sentinel/utils"
	model"github.com/negeek/solar-sphere/solar-sentinel/repository/v1"

)

// TODO Only Admins should access this endpoint.
func CreateDeviceID(w http.ResponseWriter, r *http.Request){
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

	if device.ID == "" {
		device.ID=utils.GenerateID()
	}
	
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

func DownloadSolarIrrData(w http.ResponseWriter, r *http.Request){
	var (
		device = &model.Device{}
		err error
		data []model.SolarIrradiance
		headers []string
		dataSample = model.SolarIrradiance{}
		file os.File
	)

	// param
	vars := mux.Vars(r)
	device.DeviceID = vars["device_id"]


	// Prepare CSV file
	filename:= device.DeviceID+".csv"
	file, err = os.Create(filename)
	if err != nil {
		utils.JsonResponse(w, false, http.StatusBadRequest , "Failed to create csv file", nil)
		return
	}
	defer file.Close()

	data, err = device.GetAllSolarData()
	if err != nil{
		utils.JsonResponse(w, false, http.StatusBadRequest , err.Error(), nil)
		return	
	}

	// Prepare CSV data
	csvData := [][]string{{"DeviceID", "DateUpdated"}}

	for _, solar := range data {
		row := []string{solar.DeviceID, solar.DateUpdated.Format(time.RFC3339)}
		for key := range solar.Data {
			row = append(row, key)
		}
		csvData = append(csvData, row)
	}

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, row := range csvData {
		err := writer.Write(row)
		if err != nil {
			utils.JsonResponse(w, false, http.StatusBadRequest , err.Error(), nil)
			return
		}
	}

	// Serve the CSV file for download
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	http.ServeFile(w, r, filename)
}