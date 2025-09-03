// MongoDB initialization script
db = db.getSiblingDB('Solana');

// Create collections with validation
db.createCollection('licenses', {
   validator: {
      $jsonSchema: {
         bsonType: "object",
         required: ["apiKey", "isValid", "usageCount", "createdAt"],
         properties: {
            apiKey: {
               bsonType: "string",
               description: "must be a string and is required"
            },
            isValid: {
               bsonType: "bool",
               description: "must be a boolean and is required"
            },
            usageCount: {
               bsonType: "int",
               minimum: 0,
               description: "must be an integer >= 0 and is required"
            },
            maxUsage: {
               bsonType: "int",
               minimum: 0,
               description: "must be an integer >= 0"
            },
            createdAt: {
               bsonType: "date",
               description: "must be a date and is required"
            },
            updatedAt: {
               bsonType: "date",
               description: "must be a date"
            }
         }
      }
   }
});

// Create indexes for performance
db.licenses.createIndex({ "apiKey": 1 }, { unique: true });
db.licenses.createIndex({ "isValid": 1 });
db.licenses.createIndex({ "createdAt": 1 });

// Insert sample license for testing (optional)
db.licenses.insertOne({
   apiKey: "test-api-key-12345",
   isValid: true,
   usageCount: 0,
   maxUsage: 1000,
   createdAt: new Date(),
   updatedAt: new Date()
});

print("MongoDB initialization completed successfully!");
print("Created Solana database with licenses collection");
print("Added sample test license: test-api-key-12345");