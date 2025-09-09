const { generateToken } = require('../middleware/auth');
require('dotenv').config();

class AuthController {
  // Login endpoint for JWT token generation
  static async login(req, res) {
    try {
      const { username, password } = req.body;
      
      // Simple admin authentication (in production, use proper user management)
      const adminUsername = process.env.ADMIN_USERNAME || 'admin';
      const adminPassword = process.env.ADMIN_PASSWORD || 'admin123!';
      
      if (username === adminUsername && password === adminPassword) {
        const token = generateToken({
          username: adminUsername,
          role: 'admin',
          iat: Math.floor(Date.now() / 1000)
        });
        
        res.json({
          success: true,
          message: 'Login successful',
          data: {
            token,
            user: {
              username: adminUsername,
              role: 'admin'
            },
            expires_in: process.env.JWT_EXPIRES_IN || '24h'
          }
        });
      } else {
        res.status(401).json({
          success: false,
          message: 'Invalid credentials'
        });
      }
    } catch (error) {
      console.error('Error during login:', error);
      res.status(500).json({
        success: false,
        message: 'Internal server error',
        error: process.env.NODE_ENV === 'development' ? error.message : undefined
      });
    }
  }
  
  // Verify token endpoint
  static async verifyToken(req, res) {
    try {
      // If we reach here, the token is valid (middleware already verified it)
      res.json({
        success: true,
        message: 'Token is valid',
        data: {
          user: req.user,
          valid: true
        }
      });
    } catch (error) {
      console.error('Error verifying token:', error);
      res.status(500).json({
        success: false,
        message: 'Internal server error',
        error: process.env.NODE_ENV === 'development' ? error.message : undefined
      });
    }
  }
  
  // Get API information
  static async getApiInfo(req, res) {
    try {
      res.json({
        success: true,
        message: 'API information retrieved successfully',
        data: {
          name: 'FreeRADIUS API',
          version: '1.0.0',
          description: 'REST API for FreeRADIUS management with NAS and user CRUD operations',
          endpoints: {
            authentication: {
              login: 'POST /api/v1/auth/login',
              verify: 'GET /api/v1/auth/verify'
            },
            nas: {
              list: 'GET /api/v1/nas',
              get: 'GET /api/v1/nas/:id',
              create: 'POST /api/v1/nas',
              update: 'PUT /api/v1/nas/:id',
              delete: 'DELETE /api/v1/nas/:id',
              stats: 'GET /api/v1/nas/stats'
            },
            users: {
              list: 'GET /api/v1/users',
              get: 'GET /api/v1/users/:username',
              create: 'POST /api/v1/users',
              update: 'PUT /api/v1/users/:username',
              delete: 'DELETE /api/v1/users/:username',
              stats: 'GET /api/v1/users/stats',
              attributes: 'GET /api/v1/users/:username/attributes',
              reply_attributes: 'GET /api/v1/users/:username/reply-attributes',
              add_attribute: 'POST /api/v1/users/:username/attributes',
              remove_attribute: 'DELETE /api/v1/users/:username/attributes'
            }
          },
          authentication_methods: [
            'JWT Token (Bearer)',
            'API Key (X-API-Key header)'
          ],
          supported_formats: ['JSON'],
          rate_limiting: {
            window: '15 minutes',
            max_requests: 100
          }
        }
      });
    } catch (error) {
      console.error('Error getting API info:', error);
      res.status(500).json({
        success: false,
        message: 'Internal server error',
        error: process.env.NODE_ENV === 'development' ? error.message : undefined
      });
    }
  }
  
  // Health check endpoint
  static async healthCheck(req, res) {
    try {
      const db = require('../config/database');
      
      // Test database connection
      await db.query('SELECT 1');
      
      res.json({
        success: true,
        message: 'API is healthy',
        data: {
          status: 'healthy',
          timestamp: new Date().toISOString(),
          uptime: process.uptime(),
          database: 'connected',
          memory_usage: process.memoryUsage(),
          node_version: process.version
        }
      });
    } catch (error) {
      console.error('Health check failed:', error);
      res.status(503).json({
        success: false,
        message: 'API is unhealthy',
        data: {
          status: 'unhealthy',
          timestamp: new Date().toISOString(),
          database: 'disconnected',
          error: error.message
        }
      });
    }
  }
}

module.exports = AuthController;