const express = require('express');
const redis = require('redis');
const app = express();

// Initialize the Redis client
const client = redis.createClient({
  url: `redis://:${process.env.password}@${process.env.host}:${process.env.port}`
});

client.on('error', (err) => console.log('Redis Client Error', err));

// Connect the Redis client
(async () => {
  try {
    await client.connect();
    console.log("Redis client connected");
  } catch (error) {
    console.error("Failed to connect to Redis", error);
  }
})();

// Handle Redis client errors
client.on('error', (err) => {
  console.error('Redis client error', err);
});

app.get('/hello1', async (req, res) => {
  try {
    await writeSession();
    res.send("Hello-1");
  } catch (error) {
    res.status(500).send("Internal Server Error");
    console.error("Error in /hello1 route:", error);
  }
});

app.listen(8000, () => console.log('Hello-1 listening on port 8000!'));

// Write session to Redis
async function writeSession() {
  const dt = new Date();
  const utcDate = dt.toUTCString();

  try {
    // Check and increment the session ID
    let id = await client.get('latest_session_id');
    id = id == null ? 1 : Number(id) + 1;

    // Store the updated session ID and the timestamp
    await client.set('latest_session_id', id.toString());
    await client.set(id.toString(), utcDate);
    
    console.log("Session ID:", id);
  } catch (error) {
    console.error("Error in writeSession:", error);
    throw error;  // Rethrow to let the route handle the error
  }
}