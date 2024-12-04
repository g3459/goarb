import "./router.sol";

import {IPoolFactory as IVeloV2Factory} from "./interfaces/velodrome-finance/contracts/contracts/interfaces/factories/IPoolFactory.sol";
import {IPool as IVeloV2Pool} from "./interfaces/velodrome-finance/contracts/contracts/interfaces/IPool.sol";
import {ICLFactory as IVeloV3Factory} from "./interfaces/velodrome-finance/slipstream/contracts/core/interfaces/ICLFactory.sol";
import {ICLPool as IVeloV3Pool} from "./interfaces/velodrome-finance/slipstream/contracts/core/interfaces/ICLPool.sol";
import "./interfaces/Uniswap/v2-core/contracts/interfaces/IUniswapV2Pair.sol";
import "./interfaces/Uniswap/v2-core/contracts/interfaces/IUniswapV2Factory.sol";
import "./interfaces/Uniswap/v3-core/contracts/interfaces/IUniswapV3Pool.sol";
import "./interfaces/Uniswap/v3-core/contracts/interfaces/IUniswapV3Factory.sol";
import "./interfaces/cryptoalgebra/Algebra/src/core/contracts/interfaces/IAlgebraFactory.sol";
import "./interfaces/cryptoalgebra/Algebra/src/core/contracts/interfaces/IAlgebraPool.sol";
import "./interfaces/openzeppelin/openzeppelin-contracts/contracts/token/ERC20/IERC20.sol";

contract CPoolFinder {
    uint256 internal constant STATE_MASK = 0x7fffffff00000000000000000000000000000000000000000000000000000000;
    uint256 internal constant PID_MASK = 0x0000000000000000000000ff0000000000000000000000000000000000000000;
    uint256 internal constant UNIV2_PID = 0x010000000000000000000000000000000000000000;
    uint256 internal constant UNIV3_PID = 0;
    uint256 internal constant UNIV3_PK_PID = 0x050000000000000000000000000000000000000000;
    uint256 internal constant UNIV3_AL_PID = 0x060000000000000000000000000000000000000000;
    uint256 internal constant ALGB_PID = 0x020000000000000000000000000000000000000000;
    uint256 internal constant VELOV2_PID = 0x030000000000000000000000000000000000000000;
    uint256 internal constant VELOV3_PID = 0x040000000000000000000000000000000000000000;

    constructor() {
        require(CRouter.FRP == false && CRouter.GPE == false);
    }

    function findPools(
        uint256 minLiqEth,
        address[] calldata tokens,
        uint256[] calldata protocols
    ) public view returns (bytes[][] memory pools) {
        unchecked {
            pools = new bytes[][](tokens.length);
            for (uint256 t0; t0 < tokens.length; t0++) {
                for (uint256 t1; t1 < tokens.length; t1++) {
                    if (t0 == t1 || tokens[t0] > tokens[t1]) {
                        continue;
                    }
                    bytes memory _pools = findPoolsSingle(tokens[t0], tokens[t1], protocols);
                    if (_pools.length > 0) {
                        if (pools[t0].length == 0) {
                            pools[t0] = new bytes[](tokens.length);
                        }
                        pools[t0][t1] = _pools;
                    }
                }
            }
            (uint256[] memory amounts, ) = CRouter.findRoutesInt(2, 0, minLiqEth, pools);
            filterPools(amounts, pools);
        }
    }

    function findPoolsSingle(
        address t0,
        address t1,
        uint256[] calldata protocols
    ) public view returns (bytes memory pools) {
        unchecked {
            if (t0 > t1) {
                (t0, t1) = (t1, t0);
            }
            assembly {
                pools := mload(0x40)
                mstore(0x40, add(pools, 0x20))
            }
            for (uint256 i; i < protocols.length; i++) {
                address factory = address(uint160(protocols[i]));
                uint256 pid = protocols[i] & PID_MASK;
                if (pid == UNIV3_PID) {
                    mstoreUniV3Pool(factory, t0, t1, 100, 1);
                    mstoreUniV3Pool(factory, t0, t1, 500, 10);
                    mstoreUniV3Pool(factory, t0, t1, 3000, 60);
                    mstoreUniV3Pool(factory, t0, t1, 10000, 200);
                } else if (pid == UNIV2_PID) {
                    mstoreUniV2Pool(factory, t0, t1);
                } else if (pid == ALGB_PID) {
                    mstoreAlgbPool(factory, t0, t1);
                } else if (pid == VELOV2_PID) {
                    mstoreVeloV2Pool(factory, t0, t1, true);
                    mstoreVeloV2Pool(factory, t0, t1, false);
                } else if (pid == VELOV3_PID) {
                    mstoreVeloV3Pool(factory, t0, t1, 1);
                    mstoreVeloV3Pool(factory, t0, t1, 50);
                    mstoreVeloV3Pool(factory, t0, t1, 100);
                    mstoreVeloV3Pool(factory, t0, t1, 200);
                    mstoreVeloV3Pool(factory, t0, t1, 2000);
                } else if (pid == UNIV3_PK_PID) {
                    mstoreUniV3Pool(factory, t0, t1, 100, 1);
                    mstoreUniV3Pool(factory, t0, t1, 500, 10);
                    mstoreUniV3Pool(factory, t0, t1, 2500, 50);
                    mstoreUniV3Pool(factory, t0, t1, 10000, 200);
                } else if (pid == UNIV3_AL_PID) {
                    mstoreUniV3Pool(factory, t0, t1, 100, 2);
                    mstoreUniV3Pool(factory, t0, t1, 200, 4);
                    mstoreUniV3Pool(factory, t0, t1, 300, 6);
                    mstoreUniV3Pool(factory, t0, t1, 400, 8);
                    mstoreUniV3Pool(factory, t0, t1, 3000, 60);
                    mstoreUniV3Pool(factory, t0, t1, 10000, 200);
                }
            }
            uint256 len;
            assembly {
                len := sub(mload(0x40), add(pools, 0x20))
            }
            if (len == 0) {
                delete pools;
            } else {
                assembly {
                    mstore(pools, len)
                }
            }
        }
    }

    function filterPools(uint256[] memory fAmounts, bytes[][] memory pools) internal pure {
        unchecked {
            for (uint256 t0; t0 < pools.length; t0++) {
                for (uint256 t1; t1 < pools[t0].length; t1++) {
                    if (t0 == t1) {
                        continue;
                    }
                    bytes memory _pools = pools[t0][t1];
                    if (_pools.length == 0) {
                        continue;
                    }
                    uint256 _len;
                    for (uint256 p; p < _pools.length; p += 0x40) {
                        uint256 slot0;
                        assembly {
                            slot0 := mload(add(_pools, add(p, 0x20)))
                        }
                        uint256 rt0 = slot0 >> 128;
                        uint256 rt1 = uint128(slot0);
                        if (rt0 <= fAmounts[t0] && rt1 <= fAmounts[t1]) {
                            continue;
                        }
                        assembly {
                            _len := add(_len, 0x20)
                            mstore(add(_pools, _len), slot0)
                            _len := add(_len, 0x20)
                            mstore(add(_pools, _len), mload(add(_pools, add(p, 0x40))))
                        }
                    }
                    if (_len > 0) {
                        assembly {
                            mstore(_pools, _len)
                        }
                    } else {
                        delete pools[t0][t1];
                    }
                }
            }
            for (uint256 t0; t0 < pools.length; t0++) {
                bool b;
                for (uint256 t1; t1 < pools[t0].length; t1++) {
                    if (pools[t0][t1].length > 0) {
                        b = true;
                    }
                }
                if (!b) {
                    delete pools[t0];
                }
            }
        }
    }

    function mstoreUniV2Pool(
        address factory,
        address t0,
        address t1
    ) internal view {
        unchecked {
            bytes32 fmp;
            assembly {
                fmp := mload(0x40)
            }
            address pool;
            {
                bytes4 sel = IUniswapV2Factory(factory).getPair.selector;
                assembly {
                    mstore(fmp, sel)
                    mstore(add(0x04, fmp), t0)
                    mstore(add(0x24, fmp), t1)
                    pop(staticcall(gas(), factory, fmp, 0x64, fmp, 0x20))
                    pool := mload(fmp)
                }
            }
            if (pool != address(0)) {
                uint256 reserve0;
                uint256 reserve1;
                bytes32 stateHash;
                {
                    bytes4 sel = IUniswapV2Pair(pool).getReserves.selector;
                    assembly {
                        mstore(fmp, sel)
                        pop(staticcall(gas(), pool, fmp, 0x04, fmp, 0x40))
                        reserve0 := mload(fmp)
                        reserve1 := mload(add(fmp, 0x20))
                    }
                }
                if (reserve0 > 0 && reserve1 > 0) {
                    uint8 id = 1;
                    assembly {
                        stateHash := keccak256(fmp, 0x40)
                        mstore(fmp, or(shl(128, reserve0), reserve1))
                        mstore(add(fmp, 0x20), or(and(stateHash, STATE_MASK), or(shl(216, id), or(shl(160, 3000), pool))))
                        mstore(0x40, add(fmp, 0x40))
                    }
                }
            }
        }
    }

    function mstoreUniV3Pool(
        address factory,
        address t0,
        address t1,
        uint16 fee,
        uint8 s
    ) internal view {
        unchecked {
            bytes32 fmp;
            assembly {
                fmp := mload(0x40)
            }
            address pool;
            {
                bytes4 sel = IUniswapV3Factory(factory).getPool.selector;
                assembly {
                    mstore(fmp, sel)
                    mstore(add(0x04, fmp), t0)
                    mstore(add(0x24, fmp), t1)
                    mstore(add(0x44, fmp), fee)
                    pop(staticcall(gas(), factory, fmp, 0x64, fmp, 0x20))
                    pool := mload(fmp)
                }
            }
            if (pool != address(0)) {
                uint256 liquidity;
                {
                    bytes4 sel = IUniswapV3Pool(pool).liquidity.selector;
                    assembly {
                        mstore(fmp, sel)
                        pop(staticcall(gas(), pool, fmp, 0x04, fmp, 0x20))
                        liquidity := mload(fmp)
                    }
                }
                if (liquidity > 0) {
                    uint256 sqrtPX64;
                    bytes4 sel = IUniswapV3Pool(pool).slot0.selector;
                    assembly {
                        mstore(fmp, sel)
                        pop(staticcall(gas(), pool, fmp, 0x04, fmp, 0x40))
                        sqrtPX64 := shr(32, mload(fmp))
                    }
                    uint256 reserve0 = (liquidity << 64) / (sqrtPX64 + 1);
                    uint256 reserve1 = (liquidity * sqrtPX64) >> 64;
                    if (reserve0 > 0 && reserve1 > 0) {
                        uint8 id = 0;
                        assembly {
                            let t := mload(add(fmp, 0x20))
                            let stateHash := keccak256(fmp, 0x20)
                            mstore(fmp, or(shl(128, reserve0), reserve1))
                            mstore(add(fmp, 0x20), or(and(stateHash, STATE_MASK), or(shl(216, id), or(shl(200, s), or(shl(176, and(t, 0xffffff)), or(shl(160, fee), pool))))))
                            mstore(0x40, add(fmp, 0x40))
                        }
                    }
                }
            }
        }
    }

    function mstoreAlgbPool(
        address factory,
        address t0,
        address t1
    ) internal view {
        unchecked {
            bytes32 fmp;
            assembly {
                fmp := mload(0x40)
            }
            address pool;
            {
                bytes4 sel = IAlgebraFactory(factory).poolByPair.selector;
                assembly {
                    mstore(fmp, sel)
                    mstore(add(0x04, fmp), t0)
                    mstore(add(0x24, fmp), t1)
                    pop(staticcall(gas(), factory, fmp, 0x44, fmp, 0x20))
                    pool := mload(fmp)
                }
            }
            if (pool != address(0)) {
                uint256 liquidity;
                {
                    bytes4 sel = IAlgebraPool(pool).liquidity.selector;
                    assembly {
                        mstore(fmp, sel)
                        pop(staticcall(gas(), pool, fmp, 0x04, fmp, 0x20))
                        liquidity := mload(fmp)
                    }
                }
                if (liquidity > 0) {
                    uint256 sqrtPX64;
                    {
                        bytes4 sel = IAlgebraPool(pool).globalState.selector;
                        assembly {
                            mstore(fmp, sel)
                            pop(staticcall(gas(), pool, fmp, 0x04, fmp, 0x60))
                            sqrtPX64 := shr(32, mload(fmp))
                        }
                    }
                    uint256 reserve0 = (liquidity << 64) / (sqrtPX64 + 1);
                    uint256 reserve1 = (liquidity * sqrtPX64) >> 64;
                    if (reserve0 > 0 && reserve1 > 0) {
                        uint8 id = 2;
                        assembly {
                            let t := mload(add(fmp, 0x20))
                            let fee := mload(add(fmp, 0x40))
                            let stateHash := keccak256(fmp, 0x20)
                            mstore(fmp, or(shl(128, reserve0), reserve1))
                            mstore(add(fmp, 0x20), or(and(stateHash, STATE_MASK), or(shl(216, id), or(shl(200, 60), or(shl(176, and(t, 0xffffff)), or(shl(160, fee), pool))))))
                            mstore(0x40, add(fmp, 0x40))
                        }
                    }
                }
            }
        }
    }

    function mstoreVeloV2Pool(
        address factory,
        address t0,
        address t1,
        bool stable
    ) internal view {
        unchecked {
            bytes32 fmp;
            assembly {
                fmp := mload(0x40)
            }
            address pool;
            {
                bytes4 sel = IVeloV2Factory(factory).getPool.selector;
                assembly {
                    mstore(fmp, sel)
                    mstore(add(0x04, fmp), t0)
                    mstore(add(0x24, fmp), t1)
                    mstore(add(0x44, fmp), stable)
                    pop(staticcall(gas(), factory, fmp, 0x64, fmp, 0x20))
                    pool := mload(fmp)
                }
            }
            if (pool != address(0)) {
                uint256 reserve0;
                uint256 reserve1;
                bytes32 stateHash;
                uint256 fee;
                {
                    bytes4 sel = IVeloV2Factory(factory).getFee.selector;
                    assembly {
                        mstore(fmp, sel)
                        mstore(add(0x04, pool), pool)
                        mstore(add(0x24, pool), stable)
                        pop(staticcall(gas(), factory, fmp, 0x24, fmp, 0x20))
                        fee := mload(fmp)
                    }
                }
                {
                    bytes4 sel = IVeloV2Pool(pool).getReserves.selector;
                    assembly {
                        mstore(fmp, sel)
                        pop(staticcall(gas(), pool, fmp, 0x04, fmp, 0x40))
                        reserve0 := mload(fmp)
                        reserve1 := mload(add(fmp, 0x20))
                    }
                }
                if (reserve0 > 0 && reserve1 > 0) {
                    uint8 id = 3;
                    assembly {
                        stateHash := keccak256(fmp, 0x40)
                        mstore(fmp, or(shl(128, reserve0), reserve1))
                        mstore(add(fmp, 0x20), or(and(stateHash, STATE_MASK), or(shl(216, id), or(shl(160, fee), pool))))
                        mstore(0x40, add(fmp, 0x40))
                    }
                }
            }
        }
    }

    function mstoreVeloV3Pool(
        address factory,
        address t0,
        address t1,
        int24 s
    ) internal view {
        unchecked {
            bytes32 fmp;
            assembly {
                fmp := mload(0x40)
            }
            address pool;
            {
                bytes4 sel = IVeloV3Factory(factory).getPool.selector;
                assembly {
                    mstore(fmp, sel)
                    mstore(add(0x04, fmp), t0)
                    mstore(add(0x24, fmp), t1)
                    mstore(add(0x44, fmp), s)
                    pop(staticcall(gas(), factory, fmp, 0x64, fmp, 0x20))
                    pool := mload(fmp)
                }
            }
            if (pool != address(0)) {
                uint256 liquidity;
                {
                    bytes4 sel = IVeloV3Pool(pool).liquidity.selector;
                    assembly {
                        mstore(fmp, sel)
                        pop(staticcall(gas(), pool, fmp, 0x04, fmp, 0x20))
                        liquidity := mload(fmp)
                    }
                }
                if (liquidity > 0) {
                    uint256 fee;
                    {
                        bytes4 sel = IVeloV3Factory(factory).getSwapFee.selector;
                        assembly {
                            mstore(fmp, sel)
                            mstore(add(0x04, fmp), pool)
                            pop(staticcall(gas(), factory, fmp, 0x24, fmp, 0x20))
                            fee := mload(fmp)
                        }
                    }
                    uint256 sqrtPX64;
                    {
                        bytes4 sel = IVeloV3Pool(pool).slot0.selector;
                        assembly {
                            mstore(fmp, sel)
                            pop(staticcall(gas(), pool, fmp, 0x04, fmp, 0x40))
                            sqrtPX64 := shr(32, mload(fmp))
                        }
                    }
                    uint256 reserve0 = (liquidity << 64) / (sqrtPX64 + 1);
                    uint256 reserve1 = (liquidity * sqrtPX64) >> 64;
                    if (reserve0 > 0 && reserve1 > 0) {
                        uint8 id = 0;
                        assembly {
                            let t := mload(add(fmp, 0x20))
                            let stateHash := keccak256(fmp, 0x20)
                            mstore(fmp, or(shl(128, reserve0), reserve1))
                            mstore(add(fmp, 0x20), or(and(stateHash, STATE_MASK), or(shl(216, id), or(shl(200, s), or(shl(176, and(t, 0xffffff)), or(shl(160, fee), pool))))))
                            mstore(0x40, add(fmp, 0x40))
                        }
                    }
                }
            }
        }
    }
}
