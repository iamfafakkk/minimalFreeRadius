const express = require('express');
const router = express.Router();
const NasController = require('../controllers/nasController');
const { authenticate } = require('../middleware/auth');
const {
  validateNasCreate,
  validateNasUpdate,
  validateId,
  validatePagination
} = require('../middleware/validation');

// Apply authentication to all NAS routes
router.use(authenticate);

// GET /api/v1/nas - Get all NAS entries
router.get('/', validatePagination, NasController.getAll);

// GET /api/v1/nas/stats - Get NAS statistics
router.get('/stats', NasController.getStats);

// GET /api/v1/nas/:id - Get NAS by ID
router.get('/:id', validateId, NasController.getById);

// POST /api/v1/nas - Create new NAS
router.post('/', validateNasCreate, NasController.create);

// PUT /api/v1/nas/:id - Update NAS
router.put('/:id', validateId, validateNasUpdate, NasController.update);

// DELETE /api/v1/nas/:id - Delete NAS
router.delete('/:id', validateId, NasController.delete);

module.exports = router;