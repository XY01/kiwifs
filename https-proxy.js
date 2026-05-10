const http = require('http');
const https = require('https');
const fs = require('fs');
const path = require('path');

const PROXY_PORT = 3007;

const keyPath = path.join(__dirname, 'key.pem');
const certPath = path.join(__dirname, 'cert.pem');

if (!fs.existsSync(keyPath)) {
  console.log('Generating self-signed cert...');
  const { execSync } = require('child_process');
  execSync(`openssl req -x509 -newkey rsa:2048 -keyout ${keyPath} -out ${certPath} -days 365 -nodes -subj "/CN=localhost"`, { stdio: 'inherit' });
}

const options = {
  key: fs.readFileSync(keyPath),
  cert: fs.readFileSync(certPath),
};

// Route to backend
function routeRequest(req, res) {
  let backend;
  let url = req.url;
  
  if (url.startsWith('/dashboard')) {
    backend = 'localhost:3003';
    url = url === '/dashboard' ? '/' : url;
  } else if (url.startsWith('/api') || url.startsWith('/spaces') || url.startsWith('/search')) {
    backend = 'localhost:3334';
  } else {
    backend = 'localhost:3334';
  }
  
  const proxyReq = http.request(`http://${backend}${url}`, {
    method: req.method,
    headers: { ...req.headers, host: 'localhost' },
  }, (proxyRes) => {
    res.writeHead(proxyRes.statusCode, proxyRes.headers);
    proxyRes.pipe(res, { end: true });
  });
  
  req.pipe(proxyReq, { end: true });
  proxyReq.on('error', (e) => {
    res.writeHead(502);
    res.end('Bad Gateway: ' + backend);
  });
}

https.createServer(options, routeRequest).listen(PROXY_PORT, '0.0.0.0', () => {
  console.log(`HTTPS proxy running on https://0.0.0.0:${PROXY_PORT}`);
  console.log('Routes: /dashboard/* -> localhost:3015');
  console.log('        /*        -> localhost:3333');
});