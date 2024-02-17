package v1

import (
	"net/http"
	"github.com/negeek/solar-sphere/solar-auth/utils"
	model"github.com/negeek/solar-sphere/solar-auth/repository/v1"
)

func Auth(w http.ResponseWriter, r *http.Request){
	// read  request body
	if r.Method == "POST"{
		var user = &model.User{}
		err:= utils.Unmarshall(r.Body, user)
		if err != nil{
			utils.JsonResponse(w, false, http.StatusBadRequest , err.Error(), nil)
			return	
		}

		// create user
		err=user.Create()
		if err != nil{
			utils.JsonResponse(w, false, http.StatusBadRequest , err.Error(), nil)
			return	
		}

		utils.JsonResponse(w, true, http.StatusCreated ,"Successfully Joined", user)
		return	
	}
	if r.Method == "DELETE"{
		// read  request body
		var user = &model.User{}
		err:= utils.Unmarshall(r.Body, user)
		if err != nil{
			utils.JsonResponse(w, false, http.StatusBadRequest , err.Error(), nil)
			return	
		}
		// delete user
		err=user.Delete()
		if err != nil{
			utils.JsonResponse(w, false, http.StatusBadRequest , err.Error(), nil)
			return	
		}

		utils.JsonResponse(w, true, http.StatusNoContent ,"Successfully Deleted", user)
		return
	}	
}
