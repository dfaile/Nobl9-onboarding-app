const express = require('express');
const cors = require('cors');
const bodyParser = require('body-parser');

const app = express();
app.use(cors());
app.use(bodyParser.json());

app.post('/api/create-project', async (req, res) => {
  const { appID, userGroups } = req.body;
  // TODO: Integrate with Nobl9 Go SDK or API here
  try {
    // Simulate async operation
    await new Promise(resolve => setTimeout(resolve, 1000));
    // Return success
    res.json({ success: true, message: 'Project created and users assigned.' });
  } catch (err) {
    res.status(500).json({ success: false, message: 'Backend error.' });
  }
});

const PORT = process.env.PORT || 4000;
app.listen(PORT, () => {
  console.log(`Backend listening on port ${PORT}`);
}); 