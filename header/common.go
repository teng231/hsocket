package header

import (
	"encoding/json"
	"log"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func MakeId() string {
	return primitive.NewObjectID().Hex()
}

func ToId(id string) interface{} {
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Print("parse object id error:", err)
		return nil
	}
	return objId
}

func StructToMap(strt interface{}) map[string]interface{} {
	if strt == nil {
		return map[string]interface{}{}
	}
	bin, err := json.Marshal(strt)
	if err != nil {
		log.Print(err)
		return map[string]interface{}{}
	}
	out := make(map[string]interface{})
	json.Unmarshal(bin, &out)
	return out
}
