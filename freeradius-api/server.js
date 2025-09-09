const express = require('express');
const cors = require('cors');
const helmet = require('helmet');
const rateLimit = require('express-rate-limit');
const path = require('path');
const fs = require('fs');
const http = require('http');
const https = require('https');
require('dotenv').config();

// Import routes
const authRoutes = require('./src/routes/authRoutes');
const nasRoutes = require('./src/routes/nasRoutes');
const userRoutes = require('./src/routes/userRoutes');

// Import database
const db = require('./src/config/database');

// Import Swagger configuration
const { swaggerDocument, swaggerOptions, swaggerUi, customHtml } = require('./src/config/swagger');

// Create Express app
const app = express();
const PORT = parseInt(process.env.PORT, 10) || 3000;
const HTTPS_ENABLED = String(process.env.HTTPS_ENABLED).toLowerCase() === 'true';
const HTTPS_PORT = parseInt(process.env.HTTPS_PORT, 10) || 3443;
const SSL_CERT_PATH = process.env.SSL_CERT_PATH;
const SSL_KEY_PATH = process.env.SSL_KEY_PATH;
const REDIRECT_HTTP_TO_HTTPS = String(process.env.REDIRECT_HTTP_TO_HTTPS).toLowerCase() === 'true';
const API_PREFIX = process.env.API_PREFIX || '/api/v1';
const isProduction = process.env.NODE_ENV === 'production';

// Security middleware
// If behind a reverse proxy (e.g., Nginx), trust proxy to get correct protocol info
app.set('trust proxy', 1);

// Helmet security headers
app.use(
  helmet({
    contentSecurityPolicy: {
      directives: {
        defaultSrc: ["'self'"],
        styleSrc: ["'self'", "'unsafe-inline'", "https://unpkg.com", "https://cdn.jsdelivr.net"],
        scriptSrc: ["'self'", "'unsafe-inline'", "'unsafe-eval'", "https://unpkg.com", "https://cdn.jsdelivr.net"],
        imgSrc: ["'self'", "data:", "https:", "blob:"],
        connectSrc: ["'self'", "ws:", "wss:"],
        fontSrc: ["'self'", "https://unpkg.com", "https://cdn.jsdelivr.net"],
        objectSrc: ["'none'"],
        mediaSrc: ["'self'"],
        frameSrc: ["'none'"],
      },
    },
    // These headers can cause noisy console warnings on non-HTTPS origins.
    // Disable them in non-production to improve local/dev experience.
    crossOriginEmbedderPolicy: false,
    crossOriginResourcePolicy: { policy: 'cross-origin' },
    crossOriginOpenerPolicy: isProduction ? undefined : false,
    originAgentCluster: isProduction ? undefined : false,
    // Only enable HSTS when serving over HTTPS (typically via Nginx in prod)
    hsts: HTTPS_ENABLED || isProduction ? undefined : false,
    permissionsPolicy: {
      camera: [],
      microphone: [],
      geolocation: [],
      'interest-cohort': [],
      'browsing-topics': [],
      'run-ad-auction': [],
      'join-ad-interest-group': [],
      'private-state-token-redemption': [],
      'private-state-token-issuance': [],
      'private-aggregation': [],
    },
  })
);

// CORS configuration
const corsOptions = {
  origin: process.env.CORS_ORIGIN === '*' ? true : process.env.CORS_ORIGIN?.split(',') || ['http://localhost:3000'],
  methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
  allowedHeaders: ['Content-Type', 'Authorization', 'X-API-Key'],
  credentials: true
};
app.use(cors(corsOptions));

// Rate limiting
const limiter = rateLimit({
  windowMs: parseInt(process.env.RATE_LIMIT_WINDOW_MS) || 15 * 60 * 1000, // 15 minutes
  max: parseInt(process.env.RATE_LIMIT_MAX_REQUESTS) || 100, // limit each IP to 100 requests per windowMs
  message: {
    success: false,
    message: 'Too many requests from this IP, please try again later.',
    retry_after: Math.ceil((parseInt(process.env.RATE_LIMIT_WINDOW_MS) || 15 * 60 * 1000) / 1000)
  },
  standardHeaders: true, // Return rate limit info in the `RateLimit-*` headers
  legacyHeaders: false, // Disable the `X-RateLimit-*` headers
});
app.use(limiter);

// Body parsing middleware
app.use(express.json({ limit: '10mb' }));
app.use(express.urlencoded({ extended: true, limit: '10mb' }));

// Request logging middleware
app.use((req, res, next) => {
  const timestamp = new Date().toISOString();
  console.log(`${timestamp} - ${req.method} ${req.originalUrl} - IP: ${req.ip}`);
  next();
});

// Serve static files for documentation
app.use('/docs', express.static(path.join(__dirname, 'docs')));

// Serve swagger.json
app.get('/swagger.json', (req, res) => {
  res.json(swaggerDocument);
});

// Swagger UI endpoints
app.use('/api-docs', swaggerUi.serve);
app.get('/api-docs', swaggerUi.setup(swaggerDocument, swaggerOptions));

// Alternative Swagger UI with custom HTML
app.get('/swagger', (req, res) => {
  res.send(customHtml);
});

// Health check endpoint (before API prefix)
app.get('/health', (req, res) => {
  res.json({
    success: true,
    message: 'FreeRADIUS API is running',
    timestamp: new Date().toISOString(),
    uptime: process.uptime()
  });
});

// API routes
app.use(`${API_PREFIX}/auth`, authRoutes);
app.use(`${API_PREFIX}/nas`, nasRoutes);
app.use(`${API_PREFIX}/users`, userRoutes);

// Root endpoint
app.get('/', (req, res) => {
  res.json({
    success: true,
    message: 'Welcome to FreeRADIUS API',
    version: '1.0.0',
    documentation: {
      swagger_ui: 'GET /api-docs',
      swagger_ui_custom: 'GET /swagger',
      swagger_json: 'GET /swagger.json',
      api_documentation: 'GET /docs/API_DOCUMENTATION.md',
      installation_guide: 'GET /docs/INSTALLATION_GUIDE.md',
      endpoints: {
        health: 'GET /health',
        api_info: `GET ${API_PREFIX}/auth/info`,
        login: `POST ${API_PREFIX}/auth/login`,
        nas: `GET ${API_PREFIX}/nas`,
        users: `GET ${API_PREFIX}/users`
      },
      authentication: {
        jwt: 'Use Bearer token in Authorization header',
        api_key: 'Use X-API-Key header'
      }
    }
  });
});

// 404 handler
app.use('*', (req, res) => {
  res.status(404).json({
    success: false,
    message: 'Endpoint not found',
    path: req.originalUrl,
    method: req.method,
    available_endpoints: {
      health: 'GET /health',
      api_info: `GET ${API_PREFIX}/auth/info`,
      documentation: 'GET /'
    }
  });
});

// Global error handler
app.use((err, req, res, next) => {
  console.error('Global error handler:', err);
  
  // Handle specific error types
  if (err.type === 'entity.parse.failed') {
    return res.status(400).json({
      success: false,
      message: 'Invalid JSON in request body'
    });
  }
  
  if (err.type === 'entity.too.large') {
    return res.status(413).json({
      success: false,
      message: 'Request body too large'
    });
  }
  
  // Default error response
  res.status(err.status || 500).json({
    success: false,
    message: err.message || 'Internal server error',
    error: process.env.NODE_ENV === 'development' ? err.stack : undefined
  });
});

// Graceful shutdown handler
process.on('SIGTERM', async () => {
  console.log('SIGTERM received, shutting down gracefully...');
  
  // Close database connections
  try {
    await db.close();
    console.log('Database connections closed');
  } catch (error) {
    console.error('Error closing database connections:', error);
  }
  
  process.exit(0);
});

process.on('SIGINT', async () => {
  console.log('SIGINT received, shutting down gracefully...');
  
  // Close database connections
  try {
    await db.close();
    console.log('Database connections closed');
  } catch (error) {
    console.error('Error closing database connections:', error);
  }
  
  process.exit(0);
});

// Start HTTP/HTTPS servers
const startServers = () => {
  // Always start HTTP server on PORT
  const httpServer = http.createServer(app);
  httpServer.listen(PORT, () => {
    console.log('═══════════════════════════════════════════════════════════════');
    console.log('                    FreeRADIUS API Server');
    console.log('═══════════════════════════════════════════════════════════════');
    console.log(`🚀 HTTP server running on port ${PORT}`);
    console.log(`🌐 API Base URL: http://localhost:${PORT}${API_PREFIX}`);
    console.log(`📚 API Root: http://localhost:${PORT}/`);
    console.log(`📖 Swagger UI: http://localhost:${PORT}/api-docs`);
    console.log(`📖 Swagger UI (Custom): http://localhost:${PORT}/swagger`);
    console.log(`📄 Swagger JSON: http://localhost:${PORT}/swagger.json`);
    console.log(`❤️  Health Check: http://localhost:${PORT}/health`);
    console.log(`🔐 Environment: ${process.env.NODE_ENV || 'development'}`);
    if (HTTPS_ENABLED) {
      console.log(`🔒 HTTPS also enabled on port ${HTTPS_PORT}`);
    }
    console.log('═══════════════════════════════════════════════════════════════');
  });

  if (HTTPS_ENABLED) {
    try {
      if (!SSL_CERT_PATH || !SSL_KEY_PATH) {
        console.warn('⚠️  HTTPS enabled but SSL_CERT_PATH or SSL_KEY_PATH is not set. Skipping HTTPS server.');
        return;
      }
      const httpsOptions = {
        cert: fs.readFileSync(SSL_CERT_PATH),
        key: fs.readFileSync(SSL_KEY_PATH),
      };
      const httpsServer = https.createServer(httpsOptions, app);
      httpsServer.listen(HTTPS_PORT, () => {
        console.log(`🔒 HTTPS server running on port ${HTTPS_PORT}`);
        console.log(`🌐 API Base URL: https://localhost:${HTTPS_PORT}${API_PREFIX}`);
        console.log(`📖 Swagger UI: https://localhost:${HTTPS_PORT}/api-docs`);
        console.log(`📄 Swagger JSON: https://localhost:${HTTPS_PORT}/swagger.json`);
      });

      if (REDIRECT_HTTP_TO_HTTPS) {
        // Redirect all HTTP traffic to HTTPS
        app.use((req, res, next) => {
          const isSecure = req.secure || req.headers['x-forwarded-proto'] === 'https';
          if (!isSecure) {
            const host = req.headers.host ? req.headers.host.split(':')[0] : 'localhost';
            return res.redirect(301, `https://${host}:${HTTPS_PORT}${req.originalUrl}`);
          }
          next();
        });
      }
    } catch (e) {
      console.error('❌ Failed to start HTTPS server:', e.message);
    }
  }
};

startServers();

module.exports = app;
