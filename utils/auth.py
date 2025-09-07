"""
密钥认证模块
提供密钥验证、存储、过期检查等功能
"""

import os
import json
import hashlib
import time
import yaml
from datetime import datetime, timedelta
from typing import Optional, Tuple, Dict
from .config import get_config, set_config


class AuthManager:
    """密钥认证管理器"""
    
    def __init__(self):
        self.config_key = "auth_info"
        self.used_keys_key = "used_keys"  # 已使用密钥存储键
        self.default_validity_days = 7  # 默认密钥有效期7天
        
    def _hash_key(self, key: str) -> str:
        """对密钥进行哈希处理"""
        return hashlib.sha256(key.encode('utf-8')).hexdigest()
    
    def _get_current_timestamp(self) -> float:
        """获取当前时间戳"""
        return time.time()
    
    def _is_key_used(self, key: str) -> bool:
        """检查密钥是否已被使用"""
        used_keys = get_config(self.used_keys_key) or []
        key_hash = self._hash_key(key)
        return key_hash in used_keys
    
    def _mark_key_as_used(self, key: str) -> bool:
        """标记密钥为已使用"""
        used_keys = get_config(self.used_keys_key) or []
        key_hash = self._hash_key(key)
        
        if key_hash not in used_keys:
            used_keys.append(key_hash)
            return set_config(self.used_keys_key, used_keys)
        
        return True  # 已经存在，返回true
    
    def _get_used_keys_count(self) -> int:
        """获取已使用密钥数量"""
        used_keys = get_config(self.used_keys_key) or []
        return len(used_keys)
    
    def _is_expired(self, expire_time: float) -> bool:
        """检查是否过期"""
        return self._get_current_timestamp() > expire_time
    
    def validate_key_format(self, key: str) -> bool:
        """验证密钥格式（可根据需求自定义）"""
        if not key:
            return False
        # 简单的格式验证：长度大于8位，包含字母和数字
        if len(key) < 8:
            return False
        has_letter = any(c.isalpha() for c in key)
        has_digit = any(c.isdigit() for c in key)
        return has_letter and has_digit
    
    def authenticate_key(self, key: str) -> Tuple[bool, str]:
        """
        验证密钥（支持一次性使用）
        
        Args:
            key (str): 用户输入的密钥
            
        Returns:
            Tuple[bool, str]: (验证结果, 错误信息)
        """
        if not self.validate_key_format(key):
            return False, "密钥格式不正确"
        
        # 检查密钥是否已被使用
        if self._is_key_used(key):
            return False, "该密钥已被使用，无法重复使用"
        
        # 这里可以添加服务器验证逻辑
        # 目前使用预设的有效密钥列表进行本地验证
        valid_keys = self._get_valid_keys()
        
        if key not in valid_keys:
            return False, "密钥无效"
        
        return True, ""
    
    def _get_valid_keys(self) -> list:
        """从 api.yaml 文件获取有效密钥列表"""
        api_yaml_path = ""
        try:
            # 使用 resource_path 函数获取正确的资源路径，支持打包环境
            from utils.resource import resource_path
            api_yaml_path = resource_path('api.yaml')
            
            # 检查文件是否存在
            if not os.path.exists(api_yaml_path):
                print(f"警告: api.yaml 文件不存在于 {api_yaml_path}")
                return self._get_fallback_keys()
            
            # 读取 YAML 文件
            with open(api_yaml_path, 'r', encoding='utf-8') as file:
                config = yaml.safe_load(file)
            
            # 获取有效密钥列表
            if config and 'auth' in config and 'valid_keys' in config['auth']:
                keys = config['auth']['valid_keys']
                if isinstance(keys, list) and keys:
                    print(f"成功从 api.yaml 加载 {len(keys)} 个有效密钥")
                    return keys
                else:
                    print("警告: api.yaml 中的 valid_keys 格式不正确或为空")
                    return self._get_fallback_keys()
            else:
                print("警告: api.yaml 中缺少 auth.valid_keys 配置")
                return self._get_fallback_keys()
                
        except yaml.YAMLError as e:
            print(f"错误: 解析 api.yaml 文件失败 - {str(e)}")
            return self._get_fallback_keys()
        except FileNotFoundError:
            print(f"错误: 找不到 api.yaml 文件 - {api_yaml_path}")
            return self._get_fallback_keys()
        except Exception as e:
            print(f"错误: 读取 api.yaml 文件时发生未知错误 - {str(e)}")
            return self._get_fallback_keys()
    
    def _get_fallback_keys(self) -> list:
        """获取备用密钥列表（当 api.yaml 文件无法读取时使用）"""
        print("使用备用密钥列表")
        return [
            "demo123456",
            "test987654",
            "auth456789",
            "key1234567",
        ]
    
    def save_auth_info(self, key: str, validity_days: Optional[int] = None) -> bool:
        """
        保存认证信息（使用密钥后标记为已使用）
        
        Args:
            key (str): 密钥
            validity_days (int): 有效期天数，默认使用类属性设置
            
        Returns:
            bool: 保存是否成功
        """
        if validity_days is None:
            validity_days = self.default_validity_days
            
        current_time = self._get_current_timestamp()
        expire_time = current_time + (validity_days * 24 * 60 * 60)  # 转换为秒
        
        auth_info = {
            "key_hash": self._hash_key(key),
            "activate_time": current_time,
            "expire_time": expire_time,
            "validity_days": validity_days,
            "used_key_hash": self._hash_key(key),  # 记录使用的密钥哈希
        }
        
        # 保存认证信息
        auth_saved = set_config(self.config_key, auth_info)
        
        # 标记密钥为已使用
        key_marked = self._mark_key_as_used(key)
        
        return auth_saved and key_marked
    
    def check_auth_status(self) -> Tuple[bool, str, Optional[Dict]]:
        """
        检查当前认证状态
        
        Returns:
            Tuple[bool, str, Optional[Dict]]: (是否已认证, 状态信息, 认证详情)
        """
        auth_info = get_config(self.config_key)
        
        if not auth_info:
            return False, "未找到认证信息，请输入密钥", None
        
        if not isinstance(auth_info, dict):
            return False, "认证信息格式错误，请重新输入密钥", None
        
        required_fields = ["key_hash", "activate_time", "expire_time"]
        if not all(field in auth_info for field in required_fields):
            return False, "认证信息不完整，请重新输入密钥", None
        
        # 检查是否过期
        if self._is_expired(auth_info["expire_time"]):
            expire_date = datetime.fromtimestamp(auth_info["expire_time"]).strftime("%Y-%m-%d %H:%M:%S")
            return False, f"密钥已于 {expire_date} 过期，请重新输入新密钥", auth_info
        
        # 计算剩余时间
        remaining_seconds = auth_info["expire_time"] - self._get_current_timestamp()
        remaining_days = int(remaining_seconds / (24 * 60 * 60))
        
        activate_date = datetime.fromtimestamp(auth_info["activate_time"]).strftime("%Y-%m-%d %H:%M:%S")
        expire_date = datetime.fromtimestamp(auth_info["expire_time"]).strftime("%Y-%m-%d %H:%M:%S")
        
        status_msg = f"密钥有效，剩余 {remaining_days} 天（激活时间：{activate_date}，到期时间：{expire_date}）"
        
        return True, status_msg, auth_info
    
    def verify_saved_key(self, input_key: str) -> bool:
        """
        验证输入的密钥是否与保存的密钥匹配
        
        Args:
            input_key (str): 用户输入的密钥
            
        Returns:
            bool: 是否匹配
        """
        auth_info = get_config(self.config_key)
        if not auth_info or "key_hash" not in auth_info:
            return False
        
        return auth_info["key_hash"] == self._hash_key(input_key)
    
    def clear_auth_info(self) -> bool:
        """清除认证信息"""
        return set_config(self.config_key, None)
    
    def get_auth_details(self) -> Optional[Dict]:
        """获取认证详情"""
        auth_info = get_config(self.config_key)
        if not auth_info:
            return None
        
        # 添加可读的时间信息
        if "activate_time" in auth_info:
            auth_info["activate_date"] = datetime.fromtimestamp(auth_info["activate_time"]).strftime("%Y-%m-%d %H:%M:%S")
        
        if "expire_time" in auth_info:
            auth_info["expire_date"] = datetime.fromtimestamp(auth_info["expire_time"]).strftime("%Y-%m-%d %H:%M:%S")
            remaining_seconds = auth_info["expire_time"] - self._get_current_timestamp()
            auth_info["remaining_days"] = max(0, int(remaining_seconds / (24 * 60 * 60)))
        
        return auth_info
    
    def clear_used_keys(self) -> bool:
        """清除所有已使用的密钥记录（仅供管理员使用）"""
        return set_config(self.used_keys_key, [])
    
    def get_used_keys_info(self) -> Dict:
        """获取已使用密钥的统计信息"""
        used_keys = get_config(self.used_keys_key) or []
        return {
            "total_used": len(used_keys),
            "used_keys_hashes": used_keys[:5] if len(used_keys) <= 5 else used_keys[:5] + ["..."],  # 只显示前5个
            "has_more": len(used_keys) > 5
        }
    
    def is_first_time_setup(self) -> bool:
        """检查是否是首次设置（既没有认证信息也没有使用过任何密钥）"""
        auth_info = get_config(self.config_key)
        used_keys = get_config(self.used_keys_key) or []
        return not auth_info and len(used_keys) == 0