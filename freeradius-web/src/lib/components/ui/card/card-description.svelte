<script lang="ts">
	import { cn } from '$lib/utils.js';
	let klass: string | undefined = undefined;
	export { klass as class };
</script>

<p class={cn('text-sm text-muted-foreground', klass)} {...$$restProps}>
	<slot />
</p>
