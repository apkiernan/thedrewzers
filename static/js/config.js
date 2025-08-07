// API Configuration
// This file should be updated with the actual API endpoint after deployment
window.API_CONFIG = {
    // Update this with your actual API Gateway URL or CloudFront URL
    apiEndpoint: window.location.origin + '/api',
    
    // For local development, you might want to use:
    // apiEndpoint: 'http://localhost:8080/api',
    
    // After deployment, this could be:
    // apiEndpoint: 'https://your-cloudfront-domain.cloudfront.net/api',
};