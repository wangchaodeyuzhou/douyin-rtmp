import secrets
import hashlib
import time

def generate_api_key(prefix: str = "api", length: int = 32) -> str:
    """
    生成一个安全的 API Key
    :param prefix: API key 前缀（可选，比如区分项目）
    :param length: 随机部分长度
    :return: API Key 字符串
    """
    # 使用 secrets 生成安全的随机字符串
    random_bytes = secrets.token_bytes(length)

    # 加入时间戳，避免重复
    raw_key = f"{prefix}{time.time()}".encode() + random_bytes

    # 使用 SHA256 哈希后取十六进制
    api_key = hashlib.sha256(raw_key).hexdigest()

    return f"{prefix}_{api_key}"

if __name__ == "__main__":
    prefix = 'api_'
    kk = []
    for i in range(10):
        kk.append(generate_api_key(prefix))
    print(kk)
