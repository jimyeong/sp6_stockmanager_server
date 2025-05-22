please make an updateItem api point.
the request format: {"barcode":"8801073110502","code":"105095","name":"치킨 불닭 오리지날멀티팩","type":"BOX","availableForOrder":1,"imagePath":"/uploads/1742761463173098000_Screenshot 2025-03-23 at 20.19.06.png","id":"46"}(exmaple). 
guaranteed params: 

once you've completed please give me usage and response format as well  



```
// Example client-side code for calling the saveBarcode API endpoint

// Function to save a barcode
async function saveBarcode(barcode, authToken) {
  try {
    const response = await fetch('http://your-api-url/api/v1/saveBarcode', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${authToken}` // Firebase auth token
      },
      body: JSON.stringify({
        barcode: barcode
      })
    });

    const data = await response.json();
    
    // Check if the request was successful
    if (response.ok) {
      console.log('Barcode saved successfully:', data);
      // The response will have the format:
      // {
      //   message: "Barcode saved successfully",
      //   payload: {
      //     barcode: "1234567890",
      //     message: "barcode saved"
      //   },
      //   success: true,
      //   userExists: true
      // }
      return data;
    } else {
      // Handle error responses
      console.error('Error saving barcode:', data.message);
      throw new Error(data.message || 'Failed to save barcode');
    }
  } catch (error) {
    console.error('Error saving barcode:', error);
    throw error;
  }
}

// Example usage in a React component
/*
import React, { useState } from 'react';
import { useAuth } from './auth-context'; // Your authentication context

function BarcodeScanner() {
  const [barcode, setBarcode] = useState('');
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState(null);
  const [error, setError] = useState(null);
  const { currentUser } = useAuth(); // Get the current authenticated user

  const handleScan = async (e) => {
    e.preventDefault();
    
    if (!barcode) {
      setError('Please enter a barcode');
      return;
    }

    try {
      setLoading(true);
      setError(null);
      
      // Get the current user's ID token
      const token = await currentUser.getIdToken();
      
      // Call the API
      const response = await saveBarcode(barcode, token);
      
      // Handle success
      setResult(response);
      setBarcode(''); // Clear the input
    } catch (error) {
      // Handle error
      setError(error.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <h2>Barcode Scanner</h2>
      
      <form onSubmit={handleScan}>
        <input
          type="text"
          value={barcode}
          onChange={(e) => setBarcode(e.target.value)}
          placeholder="Enter barcode"
          disabled={loading}
        />
        <button type="submit" disabled={loading}>
          {loading ? 'Saving...' : 'Save Barcode'}
        </button>
      </form>
      
      {error && <div className="error">{error}</div>}
      
      {result && (
        <div className="success">
          <p>{result.message}</p>
          <p>Barcode: {result.payload.barcode}</p>
          <p>Status: {result.payload.message}</p>
        </div>
      )}
    </div>
  );
}

export default BarcodeScanner;
*/
```
