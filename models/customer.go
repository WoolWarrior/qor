package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"

	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/roles"
)

type Customer struct {
	// ID          uuid.UUID `gorm:"primary_key;type:uuid;default:uuid_generate_v4()"`
	ID          string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time `sql:"index"`
	Name        string
	Description string
}

// DeepCopy method is to copy interface object
func DeepCopy(source interface{}, destination interface{}) {
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(source)
	json.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(&destination)
}

func ConfigureQorResourceDynamoDB(r resource.Resourcer) {
	// Configure resource with DynamoDB
	config := &aws.Config{
		Region:   aws.String("us-west-2"),
		Endpoint: aws.String("http://localhost:8000"),
	}

	// Create DynamoDB client
	svc := dynamodb.New(session.New(), config)

	customer, ok := r.(*admin.Resource)
	if !ok {
		panic(fmt.Sprintf("Unexpected resource! T: %T", r))
	}

	tableName := "Customers"

	customer.FindOneHandler = func(result interface{}, metaValues *resource.MetaValues, context *qor.Context) error {
		fmt.Println("FindOneHandler")
		if customer.HasPermission(roles.Read, context) {

			customerIDString := context.ResourceID

			// input to define the data to
			input := &dynamodb.GetItemInput{
				Key: map[string]*dynamodb.AttributeValue{
					"ID": {
						S: aws.String(customerIDString),
					},
				},
				TableName: aws.String(tableName),
			}

			resultFromDB, err := svc.GetItem(input)

			dbCustomer := Customer{}
			err = dynamodbattribute.UnmarshalMap(resultFromDB.Item, &dbCustomer)

			if err != nil {
				panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
			}

			DeepCopy(dbCustomer, &result)
			fmt.Println("Found item: ", dbCustomer)

			return err

		}

		return roles.ErrPermissionDenied
	}

	customer.FindManyHandler = func(result interface{}, context *qor.Context) error {
		fmt.Println("FindManyHandler")
		if customer.HasPermission(roles.Read, context) {

			input := &dynamodb.ScanInput{
				TableName: aws.String(tableName),
			}

			resultFromDB, err := svc.Scan(input)

			if err != nil {
				fmt.Println("Query API call failed:")
				fmt.Println((err.Error()))
				os.Exit(1)
			}

			// create a slice to store result
			dbCustomers := make([]Customer, 0)
			numResult := 0

			for _, i := range resultFromDB.Items {
				dbcustomersTMP := Customer{}
				err = dynamodbattribute.UnmarshalMap(i, &dbcustomersTMP)
				if err != nil {
					fmt.Println("Got error unmarshalling:")
					fmt.Println(err.Error())
					os.Exit(1)
				}
				dbCustomers = append(dbCustomers, dbcustomersTMP)
				numResult++

			}

			DeepCopy(dbCustomers, &result)

			fmt.Println("Found", numResult, "result(s) as below: ", dbCustomers)
			return err
		}

		return roles.ErrPermissionDenied
	}

	customer.SaveHandler = func(result interface{}, context *qor.Context) error {
		fmt.Println("SaveHandler")
		if customer.HasPermission(roles.Create, context) || customer.HasPermission(roles.Update, context) {

			var customerTMP Customer

			DeepCopy(result, &customerTMP)

			newUUID, _ := uuid.NewRandom()

			if customerTMP.ID == "" {
				customerTMP.ID = newUUID.String()
				customerTMP.CreatedAt = time.Now()
			}
			customerTMP.UpdatedAt = time.Now()

			input := &dynamodb.UpdateItemInput{
				ExpressionAttributeNames: map[string]*string{
					"#N": aws.String("Name"),
					"#D": aws.String("Description"),
					"#C": aws.String("CreatedAt"),
					"#U": aws.String("UpdatedAt:"),
				},
				ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
					":name": {
						S: aws.String(customerTMP.Name),
					},
					":description": {
						S: aws.String(customerTMP.Description),
					},
					":createdat": {
						S: aws.String(customerTMP.CreatedAt.Format(time.RFC3339)),
					},
					":updateat": {
						S: aws.String(customerTMP.UpdatedAt.Format(time.RFC3339)),
					},
				},

				Key: map[string]*dynamodb.AttributeValue{
					"ID": {
						S: aws.String(customerTMP.ID),
					},
				},
				ReturnValues:     aws.String("UPDATED_NEW"),
				TableName:        aws.String(tableName),
				UpdateExpression: aws.String("SET #N =:name, #D =:description, #C =:createdat, #U =:updateat "),
			}

			_, err := svc.UpdateItem(input)

			if err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Println("Successfully updated ", customerTMP)
			}

			return err

		}
		return roles.ErrPermissionDenied
	}

	customer.DeleteHandler = func(result interface{}, context *qor.Context) error {
		fmt.Println("DeleteHandler")
		if customer.HasPermission(roles.Delete, context) {
			// var dbCustomerTMP Customer
			// dbCustomerTMP.ID, _ = uuid.Parse(context.ResourceID)

			customerIDString := context.ResourceID

			input := &dynamodb.DeleteItemInput{
				Key: map[string]*dynamodb.AttributeValue{
					"ID": {
						S: aws.String(customerIDString),
					},
				},
				TableName: aws.String(tableName),
			}

			_, err := svc.DeleteItem(input)
			if err != nil {
				fmt.Println("Got error calling DeleteItem")
				fmt.Println(err.Error())
				return nil
			}

			fmt.Println("Deleted ", customerIDString)

			return err
		}
		return roles.ErrPermissionDenied
	}

}
