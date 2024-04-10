package v1

import (
	"time"
	"encoding/csv"
	"fmt"
	"net/http"
	"github.com/gorilla/mux"
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
		solarFields []interface{}
	)
	//  query param
	vars := mux.Vars(r)
	device.ID = vars["device_id"]

	// Prepare CSV file
	filename:= device.ID+".csv"

	// Prepare CSV data
	data, err = device.GetAllSolarData()
	if err != nil{
		utils.JsonResponse(w, false, http.StatusBadRequest , err.Error(), nil)
		return	
	}
	
	// Get fields from SolarIrradiance struct
	solarFields, err = utils.StructFieldNames(data[0])
	if err != nil{
		utils.JsonResponse(w, false, http.StatusBadRequest , err.Error(), nil)
		return
	}

	// Remove the Data field from list of SolarIrradiance struct fields
	solarFields = utils.RemoveItem(solarFields, "Data")

	// Get the fields of Data field in SolarIrradiance struct
	dataKeys:=utils.MapKeys(data[0].Data)
	headers	:= append(dataKeys, solarFields...)
	csvData := [][]interface{}{headers}

	// Get all row data to be written to csv file
	for _, solar := range data {
		dataRowValues:= utils.MapValues(solar.Data)
		dataRowValues= append(dataRowValues, solar.DeviceID)
		dataRowValues= append(dataRowValues, solar.DateUpdated.Format(time.RFC3339))
		csvData = append(csvData, dataRowValues)
	}

	// create writer and write directly to response writer
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Convert all row data to string and write to repsonse writer
	for _, row := range csvData {
		 // Convert each interface{} value to a string
		 var stringRow []string
		 for _, value := range row {
			stringValue, ok := value.(string)
			if !ok {
				// Handle if value is not a string
				stringValue = fmt.Sprintf("%v", value)
			}
			stringRow = append(stringRow, stringValue)
		 }
		err := writer.Write(stringRow)
		if err != nil {
			utils.JsonResponse(w, false, http.StatusBadRequest , err.Error(), nil)
			return
		}
	}
	
	// Set content headers and specify file name to be <device_id.csv>
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
}