const mongoHost = process.env.CONTRACTS_API_MONGODB_HOST
const mongoPort = process.env.CONTRACTS_API_MONGODB_PORT

const mongoUser = process.env.CONTRACTS_API_MONGODB_USERNAME
const mongoPassword = process.env.CONTRACTS_API_MONGODB_PASSWORD

const database = process.env.CONTRACTS_API_MONGODB_DATABASE
const collection = process.env.CONTRACTS_API_MONGODB_COLLECTION

const retrySeconds = parseInt(process.env.RETRY_CONNECTION_SECONDS || "5") || 5;

// try to connect to mongoDB until it is not available
let connection;
while(true) {
    try {
        connection = Mongo(`mongodb://${mongoUser}:${mongoPassword}@${mongoHost}:${mongoPort}`);
        break;
    } catch (exception) {
        print(`Cannot connect to mongoDB: ${exception}`);
        print(`Will retry after ${retrySeconds} seconds`)
        sleep(retrySeconds * 1000);
    }
}

// if database and collection exists, exit with success - already initialized
const databases = connection.getDBNames()
if (databases.includes(database)) {
    const dbInstance = connection.getDB(database)
    collections = dbInstance.getCollectionNames()
    if (collections.includes(collection)) {
        print(`Collection '${collection}' already exists in database '${database}'`)
        process.exit(0);
    }
}

// initialize
// create database and collection
const db = connection.getDB(database)
db.createCollection(collection)

// create indexes
db[collection].createIndex({ "contractnumber": 1 })

// insert sample data
let result = db[collection].insertMany([
    {
        "contractnumber": "ZML-2024-001",
        "name": "IT Support",
        "partner": "TechCorp s.r.o.",
        "validfrom": "2024-01-01",
        "validuntil": "2024-12-31",
        "budget": 50000,
        "status": "Active",
        "servicetype": "IT Support",
        "description": "Annual IT support and maintenance contract"
    },
    {
        "contractnumber": "ZML-2024-002",
        "name": "Cleaning Services",
        "partner": "CleanPro s.r.o.",
        "validfrom": "2024-03-01",
        "validuntil": "2025-02-28",
        "budget": 18000,
        "status": "Active",
        "servicetype": "Cleaning",
        "description": "Monthly cleaning services for all hospital floors"
    }
]);

if (result.writeError) {
    console.error(result)
    print(`Error when writing the data: ${result.errmsg}`)
}

// exit with success
process.exit(0);
