// Seed a test poll so /e/testpoll renders (for previewing the dark grid).
// Run it inside the running mongo container, against the "meet" database:
//
//   cd deploy
//   docker compose exec -T mongo mongosh "mongodb://localhost:27017/meet" < seed-test-event.js
//
// Safe to re-run (it replaces any existing "testpoll"). Creates an empty poll
// (no responses) — open it and paint your own availability to see the
// green/"if-needed" colors on the dark theme.

const shortId = "testpoll"

// 5 consecutive days starting tomorrow, each window starting at 13:00 UTC
// (≈ 9am US Eastern in summer). Absolute instants; shown in the viewer's local tz.
const base = new Date()
base.setUTCHours(13, 0, 0, 0)
base.setUTCDate(base.getUTCDate() + 1)
const dates = []
for (let i = 0; i < 5; i++) {
  dates.push(new Date(base.getTime() + i * 86400000))
}

db.events.deleteMany({ shortId })

const res = db.events.insertOne({
  shortId,
  ownerId: ObjectId("000000000000000000000000"), // guest-owned
  name: "Test poll — dark theme preview",
  duration: 8.0, // 8-hour window (≈ 9am–5pm)
  dates, // start of the window for each day (absolute UTC)
  hasSpecificTimes: false,
  daysOnly: false,
  type: "specific_dates",
  timeIncrement: NumberInt(30), // 30-minute slots (stored as int, not double)
  numResponses: NumberInt(0),
  signUpResponses: {},
})

print("Inserted test event _id=" + res.insertedId)
print("Open the dev server at:  http://localhost:8080/e/" + shortId)
