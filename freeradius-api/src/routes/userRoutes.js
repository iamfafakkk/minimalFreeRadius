const express = require('express');
const router = express.Router();
const UserController = require('../controllers/userController');
const { authenticate } = require('../middleware/auth');
const {
  validateUserCreate,
  validateUserUpdate,
  validateUsername,
  validatePagination
} = require('../middleware/validation');

// Apply authentication to all user routes
router.use(authenticate);

// GET /api/v1/users - Get all users
router.get('/', validatePagination, UserController.getAll);

// GET /api/v1/users/stats - Get user statistics
router.get('/stats', UserController.getStats);

// GET /api/v1/users/:username - Get user by username
router.get('/:username', validateUsername, UserController.getByUsername);

// GET /api/v1/users/:username/attributes - Get user attributes (radcheck)
router.get('/:username/attributes', validateUsername, UserController.getUserAttributes);

// GET /api/v1/users/:username/reply-attributes - Get user reply attributes (radreply)
router.get('/:username/reply-attributes', validateUsername, UserController.getUserReplyAttributes);

// POST /api/v1/users - Create new user
router.post('/', validateUserCreate, UserController.create);

// POST /api/v1/users/:username/attributes - Add custom attribute to user
router.post('/:username/attributes', validateUsername, UserController.addAttribute);

// PUT /api/v1/users/:username - Update user
router.put('/:username', validateUsername, validateUserUpdate, UserController.update);

// DELETE /api/v1/users/:username - Delete user
router.delete('/:username', validateUsername, UserController.delete);

// DELETE /api/v1/users/:username/attributes - Remove custom attribute from user
router.delete('/:username/attributes', validateUsername, UserController.removeAttribute);

module.exports = router;