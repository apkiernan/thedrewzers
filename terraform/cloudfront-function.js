/**
 * CloudFront Function to append .html extension to clean URLs
 * This allows /gallery to be served as /gallery.html
 */
function handler(event) {
    var request = event.request;
    var uri = request.uri;

    // Skip API routes
    if (uri.startsWith('/api/')) {
        return request;
    }

    // Skip static assets
    if (uri.startsWith('/static/')) {
        return request;
    }

    // Skip if URI already has an extension
    if (uri.includes('.')) {
        return request;
    }

    // Skip root path
    if (uri === '/' || uri === '') {
        return request;
    }

    // Append .html to clean URLs
    request.uri = uri + '.html';

    return request;
}
