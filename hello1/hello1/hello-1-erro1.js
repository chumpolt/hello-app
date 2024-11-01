const redis = require('redis');

const client = redis.createClient({
//  url: `redis://:${process.env.password}@${process.env.host}:${process.env.port}`
  url: `redis://${process.env.host}:${process.env.port}`
});

client.on('error', (err) => console.log('Redis Client Error', err));

(async () => {
  try {
    await client.connect();
    console.log("Connected to Redis");
  } catch (error) {
    console.error("Failed to connect to Redis", error);
  }
})();