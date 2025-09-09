const express = require('express');
const router = express.Router();
const AuthController = require('../controllers/authController');
const { authenticateToken } = require('../middleware/auth');
const { validateLogin } = require('../middleware/validation');

// POST /api/v1/auth/login - Login and get JWT token
router.post('/login', validateLogin, AuthController.login);

// GET /api/v1/auth/verify - Verify JWT token
router.get('/verify', authenticateToken, AuthController.verifyToken);

// GET /api/v1/auth/info - Get API information (public)
router.get('/info', AuthController.getApiInfo);

// GET /api/v1/auth/health - Health check (public)
router.get('/health', AuthController.healthCheck);

module.exports = router;