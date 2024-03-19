package v1

import (
	"net/http"
	"github.com/negeek/solar-sphere/solar-auth/utils"
	model"github.com/negeek/solar-sphere/solar-auth/repository/v1"
)

func Auth(w http.ResponseWriter, r *http.Request){
	/* 
		Sign up function for users.
	*/
	var (
		user = &model.User{}
		accessKey string
		err error
	)

	// Read  request body
	err= utils.Unmarshall(r.Body, user)
	if err != nil{
		utils.JsonResponse(w, false, http.StatusBadRequest , err.Error(), nil)
		return	
	}

	// Generate Access Key and ID
	accessKey, err= utils.GenerateAccessKey(user.Email)
	if err != nil{
		utils.JsonResponse(w, false, http.StatusBadRequest , "Error generating access key", nil)
		return	
	}
	user.ID=utils.GenerateUserID(user.Email)

	// Create User
	err=user.Create()
	if err != nil{
		utils.JsonResponse(w, false, http.StatusBadRequest , "Error creating user", nil)
		return	
	}

	// Response
	utils.JsonResponse(w, true, http.StatusCreated ,"Successfully joined", map[string]interface{}{"email":user.Email,
	"access_key":accessKey})
	return	
}

func NewAccessKey(w http.ResponseWriter, r *http.Request){
	/* 
		Function grants users new access keys but users have to provide previous access key.
	*/
	var (
		key = &model.RevokedKey{}
		err error
		claim = &utils.UserClaim{} 
		accessKey string
	)

	// Read  request body
	err= utils.Unmarshall(r.Body, key)
	if err != nil{
		utils.JsonResponse(w, false, http.StatusBadRequest , err.Error(), nil)
		return	
	}

	// Verify their access key and make sure it matches the email provided.
	claim, err = utils.VerifyAccessKey(key.Key)
	if err != nil{
		utils.JsonResponse(w, false, http.StatusBadRequest ,"Invalid access key", nil)
		return	
	}

	// Authenticate
	if claim.Email != key.Email{
		utils.JsonResponse(w, false, http.StatusBadRequest , "Invalid access key or email", nil)
		return
	}

	// Revoke previous key
	err = key.Revoke()
	if err != nil{
		utils.JsonResponse(w, false, http.StatusBadRequest , "Error Revoking Key. Ensure you haven't revoked it previously", nil)
		return	
	}

	// Generate new access key for user
	accessKey, err= utils.GenerateAccessKey(key.Email)
	if err != nil{
		utils.JsonResponse(w, false, http.StatusBadRequest , "Error generating new access key", nil)
		return	
	}

	// Response
	utils.JsonResponse(w, true, http.StatusCreated ,"Successfully changed access key", map[string]interface{}{"email":key.Email,
		"access_key":accessKey})
		return

}

func RecoverAccessKey() {

	/* 
		Function that recovers users lost access key.
		In actual sense, it generates new one for them.
		I will need to get their email, send them a link in the email. Once they click it, they get their access key in another email sent to them.
	*/

}
