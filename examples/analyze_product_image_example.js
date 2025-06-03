// Example client-side code for calling the analyzeProductImage API endpoint

// Function to convert file to base64
function fileToBase64(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.readAsDataURL(file);
    reader.onload = () => resolve(reader.result.split(',')[1]); // Split to get just the base64 part
    reader.onerror = error => reject(error);
  });
}

// Function to analyze a product image
async function analyzeProductImage(imageFile, authToken) {
  try {
    // Convert image to base64
    const base64Image = await fileToBase64(imageFile);
    
    const response = await fetch('http://your-api-url/api/v1/analyzeProductImage', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${authToken}` // Pass Firebase auth token for authentication
      },
      body: JSON.stringify({
        image: base64Image
      })
    });

    const data = await response.json();
    
    // Check if the request was successful
    if (response.ok) {
      console.log('Product analysis completed:', data);
      // The response will have the format:
      // {
      //   message: "Product analysis completed",
      //   payload: {
      //     analysis: "1. Product Name: Example Product\n2. Expiry Date: 01/01/2025\n..."
      //   },
      //   success: true,
      //   userExists: true
      // }
      return data;
    } else {
      // Handle error responses
      console.error('Error analyzing product image:', data.message);
      throw new Error(data.message || 'Failed to analyze product image');
    }
  } catch (error) {
    console.error('Error analyzing product image:', error);
    throw error;
  }
}

// Example usage in a React component
/*
import React, { useState } from 'react';
import { useAuth } from './auth-context'; // Your authentication context

function ProductScanner() {
  const [image, setImage] = useState(null);
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState(null);
  const [error, setError] = useState(null);
  const { currentUser } = useAuth(); // Get the current authenticated user
  const handleImageChange = (e) => {
    if (e.target.files[0]) {
      setImage(e.target.files[0]);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    if (!image) {
      setError('Please select an image');
      return;
    }

    try {
      setLoading(true);
      setError(null);
      
      // Get the current user's ID token
      const token = await currentUser.getIdToken();
      
      // Call the API
      const response = await analyzeProductImage(image, token);
      
      // Parse the analysis to display in a structured way
      const analysisLines = response.payload.analysis.split('\n');
      const parsedAnalysis = {
        productName: analysisLines[0].replace('1. Product Name:', '').trim(),
        expiryDate: analysisLines[1].replace('2. Expiry Date:', '').trim(),
        ingredients: analysisLines[2].replace('3. Ingredients:', '').trim(),
        alcohol: analysisLines[3].replace('4. Alcohol:', '').trim(),
        halal: analysisLines[4].replace('5. Halal:', '').trim(),
        reasoning: analysisLines[5].replace('6. Reasoning:', '').trim()
      };
      
      // Handle success
      setResult(parsedAnalysis);
    } catch (error) {
      // Handle error
      setError(error.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <h2>Product Scanner</h2>
      
      <form onSubmit={handleSubmit}>
        <div>
          <input
            type="file"
            accept="image/*"
            onChange={handleImageChange}
            disabled={loading}
          />
        </div>
        
        {image && (
          <div style={{ marginTop: '10px' }}>
            <img 
              src={URL.createObjectURL(image)} 
              alt="Selected" 
              style={{ maxWidth: '300px', maxHeight: '300px' }} 
            />
          </div>
        )}
        
        <button type="submit" disabled={loading || !image}>
          {loading ? 'Analyzing...' : 'Analyze Product'}
        </button>
      </form>
      
      {error && <div className="error">{error}</div>}
      
      {result && (
        <div className="result">
          <h3>Product Analysis</h3>
          <p><strong>Product Name:</strong> {result.productName}</p>
          <p><strong>Expiry Date:</strong> {result.expiryDate}</p>
          <p><strong>Ingredients:</strong> {result.ingredients}</p>
          <p><strong>Alcohol:</strong> {result.alcohol}</p>
          <p><strong>Halal:</strong> {result.halal}</p>
          <p><strong>Reasoning:</strong> {result.reasoning}</p>
        </div>
      )}
    </div>
  );
}

export default ProductScanner;
*/