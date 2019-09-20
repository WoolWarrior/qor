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
	ID          uuid.UUID `gorm:"primary_key;type:uuid;default:uuid_generate_v4()"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time `sql:"index"`
	Name        string
	Description string
}

type CustomerStringID struct {
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

func ConfigureQorResource(r resource.Resourcer) {
	// Configure resource with dummy Objects data structure

	var dummyCustomer1 Customer
	dummyCustomer1.ID, _ = uuid.Parse("1D50A411-4927-4812-B6D0-215E8620F68B")
	dummyCustomer1.Name = "dummy customer 1"
	dummyCustomer1.Description = "the first dummy customer"
	dummyCustomer1.CreatedAt = time.Now()
	dummyCustomer1.UpdatedAt = time.Now()

	var dummyCustomer2 Customer
	dummyCustomer2.ID, _ = uuid.Parse("0052B26D-CA72-434A-BAEF-8D047A2F9F32")
	dummyCustomer2.Name = "dummy customer 2"
	dummyCustomer2.Description = "the second dummy customer"
	dummyCustomer2.CreatedAt = time.Now()
	dummyCustomer2.UpdatedAt = time.Now()

	var dummyCustomer3 Customer
	dummyCustomer3.ID, _ = uuid.Parse("6400F6FA-56CA-457E-927B-CB18F44B298F")
	dummyCustomer3.Name = "dummy customer 3"
	dummyCustomer3.Description = "the third dummy customer"
	dummyCustomer3.CreatedAt = time.Now()
	dummyCustomer3.UpdatedAt = time.Now()

	dummyCustomers := make([]Customer, 0)
	dummyCustomers = append(dummyCustomers, dummyCustomer1)
	dummyCustomers = append(dummyCustomers, dummyCustomer2)
	dummyCustomers = append(dummyCustomers, dummyCustomer3)

	customer, ok := r.(*admin.Resource)
	if !ok {
		panic(fmt.Sprintf("Unexpected resource! T: %T", r))
	}
	// find record and decode it to result
	customer.FindOneHandler = func(result interface{}, metaValues *resource.MetaValues, context *qor.Context) error {

		if customer.HasPermission(roles.Read, context) {

			var dummyCustomerTMP Customer
			fmt.Println("result before FindOneHandler: ", result)
			dummyCustomerTMP.ID, _ = uuid.Parse(context.ResourceID)
			for i := 0; i < len(dummyCustomers); i++ {
				if dummyCustomers[i].ID == dummyCustomerTMP.ID {
					var buf bytes.Buffer
					json.NewEncoder(&buf).Encode(dummyCustomers[i])
					json.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(&result)
				}
			}

			fmt.Println("result after FindOneHandler: ", result)

			return nil
		}

		return roles.ErrPermissionDenied
	}

	customer.FindManyHandler = func(result interface{}, context *qor.Context) error {
		if customer.HasPermission(roles.Read, context) {

			var buf bytes.Buffer
			json.NewEncoder(&buf).Encode(dummyCustomers)
			json.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(&result)
			return nil
		}

		return roles.ErrPermissionDenied

	}

	customer.SaveHandler = func(result interface{}, context *qor.Context) error {
		if customer.HasPermission(roles.Create, context) || customer.HasPermission(roles.Update, context) {
			tmpUUID, _ := uuid.Parse("00000000-0000-0000-0000-000000000000")

			var dummyCustomerTMP Customer

			var buf bytes.Buffer
			json.NewEncoder(&buf).Encode(result)
			json.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(&dummyCustomerTMP)

			if dummyCustomerTMP.ID == tmpUUID {
				dummyCustomerTMP.ID, _ = uuid.NewRandom()
				dummyCustomers = append(dummyCustomers, dummyCustomerTMP)
			} else {
				for i := 0; i < len(dummyCustomers); i++ {
					if dummyCustomers[i].ID == dummyCustomerTMP.ID {
						var buf bytes.Buffer
						json.NewEncoder(&buf).Encode(dummyCustomerTMP)
						json.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(&dummyCustomers[i])
					}
				}
			}

			return nil
		}
		return roles.ErrPermissionDenied
	}

	customer.DeleteHandler = func(result interface{}, context *qor.Context) error {
		if customer.HasPermission(roles.Delete, context) {

			var dummyCustomerTMP Customer
			fmt.Println("result before DeleteHandler: ", result)
			dummyCustomerTMP.ID, _ = uuid.Parse(context.ResourceID)

			for i := 0; i < len(dummyCustomers); i++ {
				if dummyCustomers[i].ID == dummyCustomerTMP.ID {
					copy(dummyCustomers[i:], dummyCustomers[i+1:])
					dummyCustomers = dummyCustomers[:len(dummyCustomers)-1]
				}
			}

			return nil
		}
		return roles.ErrPermissionDenied
	}

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

	customer.FindOneHandler = func(result interface{}, metaValues *resource.MetaValues, context *qor.Context) error {
		fmt.Println("FindOneHandler")
		if customer.HasPermission(roles.Read, context) {

			dbcustomerTMP2 := CustomerStringID{}

			DeepCopy(dbcustomerTMP2, &result)

			// input to define the data to
			input := &dynamodb.GetItemInput{
				Key: map[string]*dynamodb.AttributeValue{
					"ID": {
						S: aws.String(context.ResourceID),
					},
				},
				TableName: aws.String("Customers"),
			}

			result2, err2 := svc.GetItem(input)
			fmt.Println(result2)
			fmt.Println(result2.Item)

			err2 = dynamodbattribute.UnmarshalMap(result2.Item, &dbcustomerTMP2)

			if err2 != nil {
				panic(fmt.Sprintf("Failed to unmarshal Record, %v", err2))
			}

			// var buf bytes.Buffer
			// json.NewEncoder(&buf).Encode(dbcustomerTMP2)
			// json.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(&result)
			DeepCopy(dbcustomerTMP2, &result)

			return err2

		}

		return roles.ErrPermissionDenied
	}

	customer.FindManyHandler = func(result interface{}, context *qor.Context) error {
		fmt.Println("FindManyHandler 2")
		if customer.HasPermission(roles.Read, context) {

			// var dbcustomers []Customer
			// err := table.Scan().All(&dbcustomers)

			// var buf bytes.Buffer
			// json.NewEncoder(&buf).Encode(dbcustomers)
			// json.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(&result)

			input := &dynamodb.ScanInput{
				TableName: aws.String("Customers"),
			}

			result2, err2 := svc.Scan(input)

			if err2 != nil {
				fmt.Println("Query API call failed:")
				fmt.Println((err2.Error()))
				os.Exit(1)
			}

			dbcustomers2 := make([]CustomerStringID, 0)

			numResult := 0
			for _, i := range result2.Items {
				dbcustomersTMP := CustomerStringID{}
				err2 = dynamodbattribute.UnmarshalMap(i, &dbcustomersTMP)
				if err2 != nil {
					fmt.Println("Got error unmarshalling:")
					fmt.Println(err2.Error())
					os.Exit(1)
				}
				dbcustomers2 = append(dbcustomers2, dbcustomersTMP)
				numResult++
				fmt.Println("number of result: ", numResult)

			}

			var buf bytes.Buffer
			json.NewEncoder(&buf).Encode(dbcustomers2)
			json.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(&result)
			fmt.Println("Got the table ")
			return err2
		}

		return roles.ErrPermissionDenied
	}

	customer.SaveHandler = func(result interface{}, context *qor.Context) error {
		fmt.Println("SaveHandler")
		if customer.HasPermission(roles.Create, context) || customer.HasPermission(roles.Update, context) {

			tmpUUID, _ := uuid.Parse("00000000-0000-0000-0000-000000000000")

			var dummyCustomerTMP Customer

			var buf bytes.Buffer
			json.NewEncoder(&buf).Encode(result)
			json.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(&dummyCustomerTMP)

			// var err error

			if dummyCustomerTMP.ID == tmpUUID {
				dummyCustomerTMP.ID, _ = uuid.NewRandom()
			}

			// type CustomerStringID struct {
			// 	ID          string
			// 	CreatedAt   time.Time
			// 	UpdatedAt   time.Time
			// 	DeletedAt   *time.Time `sql:"index"`
			// 	Name        string
			// 	Description string
			// }

			fmt.Println(dummyCustomerTMP)

			input := &dynamodb.UpdateItemInput{
				ExpressionAttributeNames: map[string]*string{
					"#N": aws.String("Name"),
					"#D": aws.String("Description"),
					// "#C": aws.String("CreatedAt:"),
					// "#U": aws.String("UpdatedAt:"),
				},
				ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
					":name": {
						S: aws.String(dummyCustomerTMP.Name),
					},
					":description": {
						S: aws.String(dummyCustomerTMP.Description),
					},
					// ":createdat": {
					// 	S: aws.String(dummyCustomerTMP.CreatedAt.String()),
					// },
					// ":updateat": {
					// 	S: aws.String(dummyCustomerTMP.UpdatedAt.String()),
					// },
				},

				Key: map[string]*dynamodb.AttributeValue{
					"ID": {
						S: aws.String(dummyCustomerTMP.ID.String()),
					},
				},
				ReturnValues:     aws.String("UPDATED_NEW"),
				TableName:        aws.String("Customers"),
				UpdateExpression: aws.String("SET #N =:name, #D =:description"),
			}

			_, err2 := svc.UpdateItem(input)

			// err = table.Put(dummyCustomerTMP).Run()
			if err2 != nil {
				fmt.Println(err2.Error())
			}

			return err2

		}
		return roles.ErrPermissionDenied
	}

	customer.DeleteHandler = func(result interface{}, context *qor.Context) error {
		fmt.Println("DeleteHandler")
		if customer.HasPermission(roles.Delete, context) {
			var dbcustomerTMP Customer
			dbcustomerTMP.ID, _ = uuid.Parse(context.ResourceID)

			// err := table.Delete("ID", dbcustomerTMP.ID).Run()

			input := &dynamodb.DeleteItemInput{
				Key: map[string]*dynamodb.AttributeValue{
					"ID": {
						S: aws.String(dbcustomerTMP.ID.String()),
					},
				},
				TableName: aws.String("Customers"),
			}

			_, err := svc.DeleteItem(input)
			if err != nil {
				fmt.Println("Got error calling DeleteItem")
				fmt.Println(err.Error())
				return nil
			}

			fmt.Println("Deleted '" + "" + "' (" + "movieYear" + ") from table " + "Customers")

			return err
		}
		return roles.ErrPermissionDenied
	}

}
