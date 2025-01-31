contract CRouter {
    bool internal constant FRP = true;
    bool internal constant GPE = true;
    uint256 internal immutable maxLen;

    uint256 internal constant STATE_MASK = 0x7fffffff00000000000000000000000000000000000000000000000000000000;
    uint256 internal constant ADDRESS_MASK = 0x000000000000000000000000ffffffffffffffffffffffffffffffffffffffff;
    uint256 internal constant DIREC_MASK = 0x8000000000000000000000000000000000000000000000000000000000000000;
    uint256 internal constant PID_MASK = 0xff000000000000000000000000000000000000000000000000000000;
    uint256 internal constant UNIV2_PID = 0x01000000000000000000000000000000000000000000000000000000;
    uint256 internal constant UNIV3_PID = 0;
    uint256 internal constant ALGB_PID = 0x02000000000000000000000000000000000000000000000000000000;
    uint256 internal constant VELOV2_PID = 0x03000000000000000000000000000000000000000000000000000000;
    uint256 internal constant UNIV2AL_PID = 0x07000000000000000000000000000000000000000000000000000000;
    uint256 internal constant UNIV2PK_PID = 0x08000000000000000000000000000000000000000000000000000000;
    uint256 internal constant UNIV3PK_PID = 0x05000000000000000000000000000000000000000000000000000000;
    uint256 internal constant UNIV3AL_PID = 0x06000000000000000000000000000000000000000000000000000000;
    uint256 internal constant VELOV3_PID = 0x04000000000000000000000000000000000000000000000000000000;
    int24 internal constant MIN_TICK = -887272;
    int24 internal constant MAX_TICK = 887272;

    constructor(uint256 _maxLen) {
        maxLen = _maxLen * 0x20;
    }

    function findRoutes(
        bytes[][] memory pools,
        uint256 amIn,
        uint8 tIn
    ) public view returns (bytes[] memory calls) {
        unchecked {
            uint256[] memory amounts = new uint256[](pools.length);
            amounts[tIn] = amIn;
            return findRoutes(pools, amounts);
        }
    }

    function findRoutes(bytes[][] memory pools, uint256[] memory amounts) internal view returns (bytes[] memory calls) {
        unchecked {
            calls = new bytes[](pools.length);
            uint256 updated = type(uint256).max >> (256 - pools.length);
            while (updated != 0) {
                for (uint256 t0; t0 < pools.length; t0++) {
                    if (updated & (1 << t0) == 0) continue;
                    updated ^= 1 << t0;
                    uint256 amIn = amounts[t0];
                    if (amIn == 0 || calls[t0].length == maxLen) continue;
                    for (uint256 t1; t1 < pools.length; t1++) {
                        if (t0 == t1) continue;
                        bytes memory _pools;
                        bool direc;
                        if (pools[t0].length > 0 && pools[t0][t1].length > 0) {
                            direc = true;
                            _pools = pools[t0][t1];
                        } else if (pools[t1].length > 0 && pools[t1][t0].length > 0) {
                            _pools = pools[t1][t0];
                        } else {
                            continue;
                        }
                        (uint256 amOut, uint256 poolCall) = quotePools(amIn, t1 == 0 ? 0 : amounts[0], direc, _pools);
                        if (amOut <= amounts[t1]) continue;
                        amounts[t1] = amOut;
                        calls[t1] = bytes.concat(calls[t0], abi.encode(poolCall));
                        updated |= 1 << t1;
                    }
                }
            }
        }
    }

    function quotePools(
        uint256 amIn,
        uint256 eth,
        bool direc,
        bytes memory _pools
    ) internal view returns (uint256 hAmOut, uint256 poolCall) {
        unchecked {
            for (uint256 p; p < _pools.length; p += 0x40) {
                uint256 slot1;
                assembly {
                    slot1 := mload(add(add(_pools, p), 0x40))
                }
                uint256 rIn;
                uint256 rOut;
                {
                    uint256 slot0;
                    assembly {
                        slot0 := mload(add(add(_pools, p), 0x20))
                    }
                    rIn = slot0 >> 128;
                    rOut = uint128(slot0);
                }
                if (!direc) (rIn, rOut) = (rOut, rIn);
                uint256 amInXFee = ((amIn - 2) * (1e6 - uint16(slot1 >> 160))) / 1e6;
                uint256 amOut = (amInXFee * rOut) / (rIn + amInXFee) - 2; ///
                if (amOut <= hAmOut) continue;
                uint256 pid = slot1 & PID_MASK;
                if (pid == UNIV3_PID || pid == ALGB_PID) {
                    int24 tl;
                    int24 tu;
                    {
                        int24 s = int24(uint24(uint16(slot1 >> 200)));
                        tl = tickLower(int24(uint24(slot1 >> 176)), s);
                        tu = tl + s;
                    }
                    if (direc ? ((rOut - amOut) << 128) / (rIn + amInXFee) < tickSqrtPX64(tl)**2 : ((rIn + amInXFee) << 128) / (rOut - amOut) > tickSqrtPX64(tu)**2) continue;
                }
                uint256 gasFee;
                if (GPE) {
                    gasFee = protGas(pid) * tx.gasprice;
                    if (eth != 0) {
                        gasFee = (amOut * gasFee) / eth;
                    }
                    if (amOut - gasFee <= hAmOut) continue;
                    if (FRP) {
                        {
                            uint256 amOutL1 = (amInXFee << 1);
                            amOutL1 = (amOutL1 * rOut) / (rIn + amOutL1);
                            if (int256(amOutL1 - gasFee) > int256((amOut - gasFee) << 1)) continue;
                        }
                        {
                            uint256 amOutR1 = (amInXFee >> 1);
                            amOutR1 = (amOutR1 * rOut) / (rIn + amOutR1);
                            if (int256(amOutR1 - gasFee) > int256((amOut - gasFee) >> 1)) continue;
                        }
                    }
                }
                hAmOut = amOut - gasFee;
                uint256 amOut56bit = compress56bit(amOut);
                poolCall = (slot1 & (STATE_MASK | PID_MASK | ADDRESS_MASK)) | (amOut56bit << 160);
            }
            if (direc) poolCall |= DIREC_MASK;
        }
    }

    function protGas(uint256 pid) internal pure returns (uint256) {
        unchecked {
            return pid == ALGB_PID ? 300000 : 150000;
        }
    }

    function compress56bit(uint256 uncompressed) internal pure returns (uint256) {
        unchecked {
            uint256 temp = uncompressed;
            uint256 rsh;
            while (uint48(temp) != temp) {
                rsh += 8;
                temp >>= 8;
            }
            temp <<= 8;
            temp |= rsh;
            return temp;
        }
    }

    function tickLower(int24 t, int24 s) internal pure returns (int24 tl) {
        unchecked {
            assembly {
                tl := mul(sub(sdiv(t, s), and(slt(t, 0), smod(t, s))), s)
            }
        }
    }

    function tickSqrtPX64(int24 tick) internal pure returns (uint256 sqrtPX64) {
        unchecked {
            if (tick < MIN_TICK) tick = MIN_TICK;
            else if (tick > MAX_TICK) tick = MAX_TICK;
            uint256 absTick;
            assembly {
                tick := signextend(2, tick)
                let mask := sar(255, tick)
                absTick := xor(mask, add(mask, tick))
            }
            uint256 price;
            assembly {
                price := xor(shl(128, 1), mul(xor(shl(128, 1), 0xfffcb933bd6fad37aa2d162d1a594001), and(absTick, 0x1)))
            }
            if (absTick & 0x2 != 0) price = (price * 0xfff97272373d413259a46990580e213a) >> 128;
            if (absTick & 0x4 != 0) price = (price * 0xfff2e50f5f656932ef12357cf3c7fdcc) >> 128;
            if (absTick & 0x8 != 0) price = (price * 0xffe5caca7e10e4e61c3624eaa0941cd0) >> 128;
            if (absTick & 0x10 != 0) price = (price * 0xffcb9843d60f6159c9db58835c926644) >> 128;
            if (absTick & 0x20 != 0) price = (price * 0xff973b41fa98c081472e6896dfb254c0) >> 128;
            if (absTick & 0x40 != 0) price = (price * 0xff2ea16466c96a3843ec78b326b52861) >> 128;
            if (absTick & 0x80 != 0) price = (price * 0xfe5dee046a99a2a811c461f1969c3053) >> 128;
            if (absTick & 0x100 != 0) price = (price * 0xfcbe86c7900a88aedcffc83b479aa3a4) >> 128;
            if (absTick & 0x200 != 0) price = (price * 0xf987a7253ac413176f2b074cf7815e54) >> 128;
            if (absTick & 0x400 != 0) price = (price * 0xf3392b0822b70005940c7a398e4b70f3) >> 128;
            if (absTick & 0x800 != 0) price = (price * 0xe7159475a2c29b7443b29c7fa6e889d9) >> 128;
            if (absTick & 0x1000 != 0) price = (price * 0xd097f3bdfd2022b8845ad8f792aa5825) >> 128;
            if (absTick & 0x2000 != 0) price = (price * 0xa9f746462d870fdf8a65dc1f90e061e5) >> 128;
            if (absTick & 0x4000 != 0) price = (price * 0x70d869a156d2a1b890bb3df62baf32f7) >> 128;
            if (absTick & 0x8000 != 0) price = (price * 0x31be135f97d08fd981231505542fcfa6) >> 128;
            if (absTick & 0x10000 != 0) price = (price * 0x9aa508b5b7a84e1c677de54f3e99bc9) >> 128;
            if (absTick & 0x20000 != 0) price = (price * 0x5d6af8dedb81196699c329225ee604) >> 128;
            if (absTick & 0x40000 != 0) price = (price * 0x2216e584f5fa1ea926041bedfe98) >> 128;
            if (absTick & 0x80000 != 0) price = (price * 0x48a170391f7dc42444e8fa2) >> 128;
            assembly {
                if sgt(tick, 0) {
                    price := div(not(0), price)
                }
                sqrtPX64 := shr(64, price)
            }
        }
    }
}
