from flask import Blueprint, request, jsonify
from app import db
from app.models import User, UserSchema
from marshmallow import ValidationError

bp = Blueprint('main', __name__)
user_schema = UserSchema()
users_schema = UserSchema(many=True)

@bp.route('/health', methods=['GET'])
def health_check():
    return jsonify({'status': 'healthy', 'service': 'microservice-app'})

@bp.route('/users', methods=['GET'])
def get_users():
    users = User.query.all()
    return jsonify(users_schema.dump(users))

@bp.route('/users', methods=['POST'])
def create_user():
    try:
        user_data = user_schema.load(request.json)
    except ValidationError as err:
        return jsonify({'errors': err.messages}), 400
    
    user = User(**user_data)
    db.session.add(user)
    db.session.commit()
    
    return jsonify(user_schema.dump(user)), 201

@bp.route('/users/<int:user_id>', methods=['GET'])
def get_user(user_id):
    user = User.query.get_or_404(user_id)
    return jsonify(user_schema.dump(user))

@bp.route('/users/<int:user_id>', methods=['PUT'])
def update_user(user_id):
    user = User.query.get_or_404(user_id)
    
    try:
        user_data = user_schema.load(request.json)
    except ValidationError as err:
        return jsonify({'errors': err.messages}), 400
    
    user.username = user_data['username']
    user.email = user_data['email']
    db.session.commit()
    
    return jsonify(user_schema.dump(user))

@bp.route('/users/<int:user_id>', methods=['DELETE'])
def delete_user(user_id):
    user = User.query.get_or_404(user_id)
    db.session.delete(user)
    db.session.commit()
    
    return '', 204