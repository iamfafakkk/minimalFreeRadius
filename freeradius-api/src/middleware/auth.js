const jwt = require('jsonwebtoken');
require('dotenv').config();

// JWT Authentication Middleware
const authenticateToken = (req, res, next) => {
  const authHeader = req.headers['authorization'];
  const token = authHeader && authHeader.split(' ')[1]; // Bearer TOKEN

  if (!token) {
    return res.status(401).json({
      success: false,
      message: 'Access token required'
    });
  }

  jwt.verify(token, process.env.JWT_SECRET, (err, user) => {
    if (err) {
      return res.status(403).json({
        success: false,
        message: 'Invalid or expired token'
      });
    }
    
    req.user = user;
    next();
  });
};

// Generate JWT Token
const generateToken = (payload) => {
  return jwt.sign(payload, process.env.JWT_SECRET, {
    expiresIn: process.env.JWT_EXPIRES_IN || '24h'
  });
};

// Admin Role Check Middleware
const requireAdmin = (req, res, next) => {
  if (!req.user || req.user.role !== 'admin') {
    return res.status(403).json({
      success: false,
      message: 'Admin access required'
    });
  }
  next();
};

// API Key Authentication (alternative to JWT)
const authenticateApiKey = (req, res, next) => {
  const apiKey = req.headers['x-api-key'];
  
  if (!apiKey) {
    return res.status(401).json({
      success: false,
      message: 'API key required'
    });
  }
  
  // In production, store API keys in database with proper hashing
  const validApiKeys = [
    process.env.API_KEY || 'freeradius-api-key-change-this'
  ];
  
  if (!validApiKeys.includes(apiKey)) {
    return res.status(403).json({
      success: false,
      message: 'Invalid API key'
    });
  }
  
  req.apiAuth = true;
  next();
};

// Flexible authentication (JWT or API Key)
const authenticate = (req, res, next) => {
  const authHeader = req.headers['authorization'];
  const apiKey = req.headers['x-api-key'];
  
  if (apiKey) {
    return authenticateApiKey(req, res, next);
  } else if (authHeader) {
    return authenticateToken(req, res, next);
  } else {
    return res.status(401).json({
      success: false,
      message: 'Authentication required (JWT token or API key)'
    });
  }
};

module.exports = {
  authenticateToken,
  authenticateApiKey,
  authenticate,
  generateToken,
  requireAdmin
};